#!/bin/bash
# SPDX-License-Identifier: MIT

set -e

if ${VERBOSE:-false}; then set -x; fi

: ${FORGEJO:=https://codeberg.org}
: ${REPO:=forgejo-integration/forgejo}
: ${TITLE:=$TAG}
: ${RELEASE_DIR:=dist/release}
: ${DOWNLOAD_LATEST:=false}
: ${TMP_DIR:=$(mktemp -d)}
: ${GNUPGHOME:=$TMP_DIR}
: ${OVERRIDE:=false}
: ${HIDE_ARCHIVE_LINK:=false}
: ${RETRY:=1}
: ${DELAY:=10}
: ${SKIP_ASSETS:=false}

RELEASE_NOTES_ASSISTANT_VERSION=v1.7.3 # renovate: datasource=forgejo-releases depName=forgejo/release-notes-assistant registryUrl=https://code.forgejo.org

TAG_FILE="$TMP_DIR/tag$$.json"
TAG_URL=$(echo "$TAG" | sed 's/\//%2F/g')

export GNUPGHOME

get_tag() {
    if ! test -f "$TAG_FILE"; then
        if api_json GET repos/$REPO/tags/"$TAG_URL" >"$TAG_FILE"; then
            echo "tag $TAG exists"
        else
            echo "tag $TAG does not exists"
        fi
    fi
    test -s "$TAG_FILE"
}

matched_tag() {
    if get_tag; then
        local sha=$(jq --raw-output .commit.sha <"$TAG_FILE")
        test "$sha" = "$SHA"
    else
        return 1
    fi
}

ensure_tag() {
    if get_tag; then
        if ! matched_tag; then
            cat "$TAG_FILE"
            echo "the tag SHA in the $REPO repository does not match the tag SHA that triggered the build: $SHA"
            return 1
        fi
    else
        create_tag
    fi
}

create_tag() {
    api_json POST repos/$REPO/tags --data-raw '{"tag_name": "'"$TAG"'", "target": "'"$SHA"'"}' >"$TAG_FILE"
}

delete_tag() {
    if get_tag; then
        api_json DELETE repos/$REPO/tags/"$TAG_URL"
        rm -f "$TAG_FILE"
    fi
}

upload_release() {
    if $PRERELEASE || echo "${TAG}" | grep -qi '\-rc'; then
        prerelease="true"
        echo "Uploading as Pre-Release"
    else
        prerelease="false"
        echo "Uploading as Stable"
    fi
    ensure_tag
    jq -n --arg title "$TITLE" --arg body "$RELEASENOTES" --arg tag "$TAG" --arg pre $prerelease '{"draft": true, "name": $title, "body": $body, "prerelease": $pre | test("true"), "tag_name": $tag }' >"$TMP_DIR"/release-payload.json
    if ${VERBOSE:-false}; then
        echo "Payload:"
        cat "$TMP_DIR"/release-payload.json | jq
    fi
    if ! api_json POST repos/$REPO/releases -d @"$TMP_DIR"/release-payload.json >"$TMP_DIR"/release.json; then
        if ${VERBOSE:-false}; then
            echo "Response:"
            cat "$TMP_DIR"/release.json | jq
        fi
        exit 1
    fi
    if [ "$SKIP_ASSETS" == 'false' ]; then
        release_id=$(jq --raw-output .id <"$TMP_DIR"/release.json)
        for file in "$RELEASE_DIR"/*; do
            # https://dev.to/pkutaj/how-to-use-jq-for-uri-encoding-2o5
            # https://unix.stackexchange.com/questions/94295/shellcheck-is-advising-not-to-use-basename-why/94307#94307
            # url encode some chars
            asset_name="$(echo -n "${file##*/}" | jq -sRr @uri)"
            if ! api POST "repos/$REPO/releases/$release_id/assets?name=$asset_name" -H "Content-Type: multipart/form-data" -F "attachment=@$file" >"$TMP_DIR/release-$asset_name.json"; then
                if ${VERBOSE:-false}; then
                    echo "Response:"
                    cat "$TMP_DIR/release-$asset_name.json" | jq
                fi
                exit 1
            fi
        done
    fi
    maybe_use_release_note_assistant
    release_draft false
}

release_draft() {
    local state="$1"

    local id=$(api_json GET repos/$REPO/releases/tags/"$TAG_URL" | jq --raw-output .id)

    api_json PATCH repos/$REPO/releases/"$id" --data-raw '{"draft": '"$state"', "hide_archive_links": '$HIDE_ARCHIVE_LINK'}'
}

maybe_use_release_note_assistant() {
    if "$RELEASE_NOTES_ASSISTANT"; then
        curl --fail -s -S -o rna https://code.forgejo.org/forgejo/release-notes-assistant/releases/download/$RELEASE_NOTES_ASSISTANT_VERSION/release-notes-assistant
        chmod +x ./rna
        mkdir -p $RELEASE_NOTES_ASSISTANT_WORKDIR
        ./rna --workdir=$RELEASE_NOTES_ASSISTANT_WORKDIR --storage release --storage-location "$TAG" --token "$TOKEN" --forgejo-url "$SCHEME://$HOST" --repository $REPO --token "$TOKEN" release "$TAG"
    fi
}

sign_release() {
    local passphrase
    if test -s "$GPG_PASSPHRASE"; then
        passphrase="--passphrase-file $GPG_PASSPHRASE"
    fi
    gpg --import --no-tty --pinentry-mode loopback $passphrase "$GPG_PRIVATE_KEY"
    for asset in "$RELEASE_DIR"/*; do
        if [[ $asset =~ .sha256$ ]]; then
            continue
        fi
        gpg --armor --detach-sign --no-tty --pinentry-mode loopback $passphrase <"$asset" >"$asset".asc
    done
}

maybe_sign_release() {
    if test -s "$GPG_PRIVATE_KEY"; then
        sign_release
    fi
}

maybe_override() {
    if test "$OVERRIDE" = "false"; then
        return
    fi
    api_json DELETE repos/$REPO/releases/tags/"$TAG_URL" >&/dev/null || true
    if get_tag && ! matched_tag; then
        delete_tag
    fi
}

upload() {
    setup_api
    maybe_sign_release
    maybe_override
    upload_release
}

setup_api() {
    if ! which jq curl; then
        apt-get -qq update
        apt-get install -y -qq jq curl
    fi
}

api_json() {
    api "$@" -H "Content-Type: application/json"
}

api() {
    method=$1
    shift
    path=$1
    shift

    curl --retry 5 --fail -X "$method" -sS -H "Authorization: token $TOKEN" "$@" $FORGEJO/api/v1/"$path"
}

wait_release() {
    local ready=false
    for i in $(seq $RETRY); do
        if api_json GET repos/$REPO/releases/tags/"$TAG_URL" | jq --raw-output .draft >"$TMP_DIR"/draft; then
            if test "$(cat "$TMP_DIR"/draft)" = "false"; then
                ready=true
                break
            fi
            echo "release $TAG is still a draft"
        else
            echo "release $TAG does not exist yet"
        fi
        echo "waiting $DELAY seconds"
        sleep $DELAY
    done
    if ! $ready; then
        echo "no release for $TAG"
        return 1
    fi
}

download() {
    setup_api
    (
        mkdir -p $RELEASE_DIR
        cd $RELEASE_DIR
        if [[ ${DOWNLOAD_LATEST} = "true" ]]; then
            echo "Downloading the latest release"
            api_json GET repos/$REPO/releases/latest >"$TMP_DIR"/assets.json
        elif [[ ${DOWNLOAD_LATEST} == "false" ]]; then
            wait_release
            echo "Downloading tagged release ${TAG}"
            api_json GET repos/$REPO/releases/tags/"$TAG_URL" >"$TMP_DIR"/assets.json
        fi
        jq --raw-output '.assets[] | "\(.browser_download_url) \(.name)"' <"$TMP_DIR"/assets.json | while read url name; do # `name` may contain whitespace, therefore, it must be last
            url=$(echo "$url" | sed "s#/download/${TAG}/#/download/${TAG_URL}/#")
            curl --fail -H "Authorization: token $TOKEN" -o "$name" -L "$url"
        done
    )
}

missing() {
    echo need upload or download argument got nothing
    exit 1
}

${@:-missing}

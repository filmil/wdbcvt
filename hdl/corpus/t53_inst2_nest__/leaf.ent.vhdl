-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: the grandchild of the nested scope cost case.
--!
--! Axis: an empty grandchild in the child

--! A leaf with one generic and nothing else.
entity leaf is
    generic (j : integer := 0);
end entity;

architecture sim of leaf is
begin
end architecture;

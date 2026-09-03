-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief The middle level of t2_hier3. Holds no signal of its own, so
--! that only the scope nesting differs from t1_hier1.

library ieee;
    use ieee.std_logic_1164.all;

entity mid is
end entity;

architecture sim of mid is
begin
    inner: entity work.leaf;
end architecture;

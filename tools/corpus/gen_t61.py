#!/usr/bin/env python3
"""Tier 61: the numbering of types under -debug all."""
import sys, os
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from gen_common import *
from gen_t60 import case, held, unlogged

C1 = "class c_t; int f = 1; endclass\n    "

if __name__ == "__main__":
    case("t61_num_cls_3f__", "a class with int, logic [3:0] and real fields",
         "class c_t; int f = 1; logic [3:0] g; real r; endclass\n    c_t h;", "h = new; h.f = 5;", "t60_dbg_class_2_", *held("h", "c_t"))
    case("t61_num_cls_rev_", "the three fields in reverse order",
         "class c_t; real r; logic [3:0] g; int f = 1; endclass\n    c_t h;", "h = new; h.f = 5;", "t61_num_cls_3f__", *held("h", "c_t"))
    case("t61_num_cls_byte", "a class with int and byte fields",
         "class c_t; int f = 1; byte b; endclass\n    c_t h;", "h = new; h.f = 5;", "t60_dbg_class___", *held("h", "c_t"))
    case("t61_num_cls_str_", "a class with a string field",
         "class c_t; int f = 1; string n; endclass\n    c_t h;", "h = new; h.f = 5;", "t60_dbg_class___", *held("h", "c_t"))
    case("t61_num_cls_q___", "a class with a queue field",
         "class c_t; int f = 1; int q[$]; endclass\n    c_t h;", "h = new; h.f = 5;", "t60_dbg_class___", *held("h", "c_t"))
    case("t61_num_q_cls___", "a queue of class handles",
         C1 + "c_t q[$];", "q.push_back(null);", "t60_dbg_queue___", *unlogged("q"))
    case("t61_num_two_cls_", "two classes with a handle each",
         "class a_t; int f = 1; endclass\n    class b_t; int g = 2; endclass\n    a_t ha;\n    b_t hb;", "ha = new; hb = new;", "t60_dbg_class_2h",
         held("ha", "a_t")[0] + held("hb", "b_t")[0], held("ha", "a_t")[1] + held("hb", "b_t")[1])
    case("t61_num_q_q_____", "a queue of queues", "int q[$][$];", "q.push_back({5});", "t60_dbg_queue___", *unlogged("q"))
    case("t61_num_q_str___", "a queue of strings", "string q[$];", 'q.push_back("a");', "t60_dbg_queue___", *unlogged("q"))
    case("t61_num_cls_int_", "an int beside a class handle",
         C1 + "c_t h;\n    int i = 7;", "h = new; i = 9;", "t60_dbg_class___",
         held("h", "c_t")[0] + [sig("tb", "i", "int", 32)], held("h", "c_t")[1] + [tr(0, "i", "7"), tr(50, "i", "9")])
    case("t61_num_cls_byti", "byte before int", "class c_t; byte b; int f = 1; endclass\n    c_t h;", "h = new; h.f = 5;", "t61_num_cls_byte", *held("h", "c_t"))
    case("t61_num_cls_long", "int and longint fields", "class c_t; int f = 1; longint l; endclass\n    c_t h;", "h = new; h.f = 5;", "t61_num_cls_byte", *held("h", "c_t"))
    case("t61_num_cls_2vec", "two vector fields", "class c_t; logic [3:0] g; logic [7:0] k; endclass\n    c_t h;", "h = new; h.g = 5;", "t60_dbg_class___", *held("h", "c_t"))
    case("t61_num_cls_2int", "two int fields", "class c_t; int f = 1; int g; endclass\n    c_t h;", "h = new; h.f = 5;", "t60_dbg_class___", *held("h", "c_t"))
    case("t61_num_cls_ibv_", "int, byte and vector fields", "class c_t; int f = 1; byte b; logic [3:0] g; endclass\n    c_t h;", "h = new; h.f = 5;", "t61_num_cls_byte", *held("h", "c_t"))
    case("t61_num_cls_cls_", "a class with a handle field", "class b_t; int g; endclass\n    class c_t; b_t hb; endclass\n    c_t h;", "h = new;", "t60_dbg_class_d_", *held("h", "c_t"))
    case("t61_num_q_vec___", "a queue of vectors", "logic [3:0] q[$];", "q.push_back(5);", "t60_dbg_queue___", *unlogged("q"))
    case("t61_num_q_byte__", "a queue of bytes", "byte q[$];", "q.push_back(5);", "t60_dbg_queue___", *unlogged("q"))
    case("t61_num_a_then_q", "an associative array then a queue", "int a[string];\n    int q[$];", 'a["k"] = 5; q.push_back(5);', "t60_dbg_assoc___",
         unlogged("a")[0] + unlogged("q")[0], [])
    case("t61_num_q_then_a", "a queue then an associative array", "int q[$];\n    int a[string];", 'q.push_back(5); a["k"] = 5;', "t61_num_a_then_q",
         unlogged("q")[0] + unlogged("a")[0], [])
    case("t61_num_ai_thn_q", "an int keyed associative array then a queue", "int a[int];\n    int q[$];", 'a[3] = 5; q.push_back(5);', "t61_num_a_then_q",
         unlogged("a")[0] + unlogged("q")[0], [])

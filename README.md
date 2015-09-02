# rsi - Robert's Scheme Interpreter

This aspirationally titled project is an experimental lisp-like
interpreter inspired by Peter Norvig's lis.py but written in
Go. Originally this was moving torwards a more Common Lisp-ish
implementation, but now it's getting to be closer to scheme as the
suggest. It's a learing process, and there is still has a long way to go.

Unlike lis.py, which used native python lists, this implementation is
based on cons lists and is aiming to be a more complete
implementation. Someday.

## Currently supports (more or less)

* integers and arithmetic
* quote
* strings (no string functions yet though)
* define (simple define)
* set!
* begin
* lambda
* if 
* cons, car, cdr

## Incomplete todo List

1. procedure defines
2. variable argument support for lambda. (also detect duplicate parameters)
3. let and friends
4. cond
5. case
6. iteration (do)
7. tail recursion
8. string functions
9. vectors
10. macros
12. proper equivalence functions
13. set-car!, set-cdr!
14. association lists
15. ports (io)
16. rationals
17. floating point


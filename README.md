# rsi - Robert's Scheme Interpreter

This aspirationally titled project is an experimental lisp-like
interpreter inspired by Peter Norvig's lis.py but written in
Go. Originally this was moving torwards a more Common Lisp-ish
implementation, but now it's getting to be closer to scheme as the
suggest. It's a learing process, and there is still has a long way to go.

Unlike lis.py, which used native python lists, this implementation is
based on cons lists and is aiming to be a more complete
implementation. Someday.

## Testing

Travis CI: https://travis-ci.org/rread/rsi

## Currently supports (more or less)

* integers and arithmetic
* quote
* strings (no string functions yet though)
* define
* set!
* begin
* lambda
* if 
* cons, car, cdr

## Incomplete todo List

- [x] procedure defines
- [ ] variable argument support for lambda. (also detect duplicate parameters)
- [ ] let and friends
- [ ] cond
- [ ] case
- [ ] iteration (do)
- [ ] tail recursion
- [ ] string functions
- [ ] vectors
- [ ] macros
- [ ] proper equivalence functions
- [ ] set-car!, set-cdr!
- [ ] association lists
- [ ] ports (io)
- [ ] rationals
- [ ] floating point


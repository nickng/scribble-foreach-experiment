# scribble-foreach-experiment

Foreach primitive experiment with Scribble Go API.

## proto

Foreach loop states (current index, number of iterations) are kept in memory
during execution as a package-scope stack. The basic algorithm is as follows:

1. The stack `fes` (foreach stack) is initialised with protocol package
1. A foreach state `s0` is entered, either
    - User calls `s0.Foreach()`, either
        * (First time) New foreach pushed on `fes` setting number of iterations, current = 1
        * (Subsequent) increment current, check that current is < number of iterations,
            - Loop body follows from `Foreach()`, when reached end of body, go back to `s0`
    - User calls `s0.EndForeach()`, check that current = number of iterations, then pop `fes`

Some potential improvements:

- The protocol package `proto` cannot be used concurrently, this should
  be changed in the final implementation such that `fes` is attached to
  the instance of protocol (e.g. initialised with `Proto.New()`)
- `HasNext` and `Foreach` are clumsy

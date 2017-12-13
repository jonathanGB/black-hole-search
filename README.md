# Black Hole Search 

## How to run
* You must have Go v1.5+
* You must install `go get github.com/fatih/color`

* Run the algorithms and print evaluation measure statistics: `go run main.go`. To see the flags available, add the flag `-help` after.
* Benchmark the algorithms: `go test -bench=.`
* Test the algorithms: `go test` or `go test -v` for more details


## Implemented Algorithms
Implementation of various algorithms for the "Black-Hole-Search" problem:
- Group [1]
- Optimal Average Time [1]
- Optimal Team Size [1]
- Optimal Time [2]
- Divide [2]

## Bibliography
1. Balamohan, Balasingham, Paola Flocchini, Ali Miri, and Nicola Santoro. "Time optimal algorithms for black hole search in rings." *Discrete Mathematics, Algorithms and Applications* 3, no. 04 (2011): 457-471. [pdf](https://pdfs.semanticscholar.org/9e74/8c8b4a9d3796cbe0de9c9777e4d223d17fdb.pdf)
2. Dobrev, Stefan, Paola Flocchini, Giuseppe Prencipe, and Nicola Santoro. "Mobile search for a black hole in an anonymous ring." *Algorithmica* 48, no. 1 (2007): 67-90. [pdf](https://pdfs.semanticscholar.org/06b1/9902ad9158c6cadf7d7882144be9c3b1fd5a.pdf)

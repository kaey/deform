package main

// Squares RNG
// https://squaresrng.wixsite.com/rand
// https://lemire.me/blog/2016/06/27/a-fast-alternative-to-the-modulo-reduction/

func rng(ctr, key uint64) uint32 {
	var x, y, z uint64

	x = ctr * key
	y = x
	z = y + key

	// round 1
	x = x*x + y
	x = (x >> 32) | (x << 32)

	// round 2
	x = x*x + z
	x = (x >> 32) | (x << 32)

	// round 3
	x = x*x + y
	x = (x >> 32) | (x << 32)

	// round 4
	return uint32((x*x + z) >> 32)
}

func rngReduce(num uint32, max uint32) uint32 {
	return uint32((uint64(num) * uint64(max)) >> 32)
}

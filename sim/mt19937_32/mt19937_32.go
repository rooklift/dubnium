package mt19937_32

/*
	The following license applies to this file:

	Copyright (C) 1997 - 2002, Makoto Matsumoto and Takuji Nishimura,
	All rights reserved.

	Redistribution and use in source and binary forms, with or without
	modification, are permitted provided that the following conditions
	are met:

	1. Redistributions of source code must retain the above copyright
	   notice, this list of conditions and the following disclaimer.

	2. Redistributions in binary form must reproduce the above copyright
	   notice, this list of conditions and the following disclaimer in the
	   documentation and/or other materials provided with the distribution.

	3. The names of its contributors may not be used to endorse or promote
	   products derived from this software without specific prior written
	   permission.

	THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
	"AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
	LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
	A PARTICULAR PURPOSE ARE DISCLAIMED.  IN NO EVENT SHALL THE COPYRIGHT
	OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
	SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
	LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
	DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
	THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
	(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
	OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/

import "fmt"

const (
	N = 624
	M = 397
	MATRIX_A = uint32(0x9908b0df)
	UPPER_MASK = uint32(0x80000000)
	LOWER_MASK = uint32(0x7fffffff)
)

var mt []uint32 = make([]uint32, N)
var mti int = N + 1

func init_genrand(s uint32) {

	mt[0] = s & uint32(0xffffffff)

	for mti = 1; mti < N; mti++ {
		mt[mti] = (uint32(1812433253) * (mt[mti - 1] ^ (mt[mti - 1] >> 30)) + uint32(mti))
		mt[mti] &= uint32(0xffffffff)		// Can likely comment this out
	}
}

func init_by_array(init_key []uint32) {

	key_length := len(init_key)

	var i, j, k int
	init_genrand(uint32(19650218))
	i = 1

	if N > key_length {
		k = N
	} else {
		k = key_length
	}

	for ; k != 0; k-- {
		mt[i] = (mt[i] ^ ((mt[i-1] ^ (mt[i-1] >> 30)) * uint32(1664525))) + init_key[j] + uint32(j)
		mt[i] &= uint32(0xffffffff)		// Can likely comment this out
		i++
		j++
		if i >= N {
			mt[0] = mt[N - 1]
			i = 1
		}
		if j >= key_length {
			j = 0
		}
	}
	for k = N - 1; k != 0; k-- {
		mt[i] = (mt[i] ^ ((mt[i - 1] ^ (mt[i - 1] >> 30)) * uint32(1566083941))) - uint32(i)
		mt[i] &= uint32(0xffffffff) 	// Can likely comment this out
		i++
		if i >= N {
			mt[0] = mt[N - 1]
			i = 1
		}
	}

	mt[0] = uint32(0x80000000)
}

func genrand_int32() uint32 {

	var y uint32
	var mag01 [2]uint32 = [2]uint32{0, MATRIX_A}

	if mti >= N {

		var kk int

		if mti == N + 1 {
			init_genrand(uint32(5489))
		}

		for kk = 0; kk < N - M ; kk++ {
			y = (mt[kk] & UPPER_MASK) | (mt[kk + 1] & LOWER_MASK)
			mt[kk] = mt[kk + M] ^ (y >> 1) ^ mag01[y & 1]
		}
		for ; kk < N - 1; kk++ {
			y = (mt[kk] & UPPER_MASK) | (mt[kk + 1] & LOWER_MASK)
			mt[kk] = mt[kk + (M - N)] ^ (y >> 1) ^ mag01[y & 1]
		}
		y = (mt[N - 1] & UPPER_MASK) | (mt[0] & LOWER_MASK)
		mt[N - 1] = mt[M - 1] ^ (y >> 1) ^ mag01[y & 1]

		mti = 0
	}

	y = mt[mti]
	mti++

	y ^= (y >> 11)
	y ^= (y << 7) & uint32(0x9d2c5680)
	y ^= (y << 15) & uint32(0xefc60000)
	y ^= (y >> 18)

	return y
}



/*

func main() {

	var init []uint32 = []uint32{0x123, 0x234, 0x345, 0x456}
	init_by_array(init)

	fmt.Printf("1000 outputs of genrand_int32()\n");
	for i := 0; i < 1000; i++ {
		fmt.Printf("%10v ", genrand_int32())
	  	if (i % 5 == 4) {
	  		fmt.Printf("\n")
	  	}
	}
}

*/

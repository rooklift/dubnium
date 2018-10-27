package mt19937_32

/*	Derived from:
	http://www.math.sci.hiroshima-u.ac.jp/~m-mat/MT/MT2002/CODES/mt19937ar.c

	The following license applies to this file:

	Copyright (c) 1997 - 2002, Makoto Matsumoto and Takuji Nishimura.
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

const (
	_N = 624
	_M = 397
	_MATRIX_A = uint32(0x9908b0df)
	_UPPER_MASK = uint32(0x80000000)
	_LOWER_MASK = uint32(0x7fffffff)
)

var mt []uint32 = make([]uint32, _N)
var mti int = _N + 1

// --------------------------------------------------------------------------------------

func Float64() float64 {
	 return genrand_real2()
}

func Uint32() uint32 {
	return genrand_int32()
}

func Seed(s uint32) {

	mt[0] = s & uint32(0xffffffff)

	for mti = 1; mti < _N; mti++ {
		mt[mti] = (uint32(1812433253) * (mt[mti - 1] ^ (mt[mti - 1] >> 30)) + uint32(mti))
		mt[mti] &= uint32(0xffffffff)	// Can likely comment this out, looks like a no-op in Go.
	}
}

func Urd() float64 {		// Mimic the behaviour of uniform_real_distribution in the C++ library I'm dealing with.
	genrand_int32()			// Discard the first call to the RNG. (Why does this work???)
	return genrand_real2()
}

// --------------------------------------------------------------------------------------

func genrand_int32() uint32 {									// [0,0xffffffff]

	var y uint32
	var mag01 [2]uint32 = [2]uint32{0, _MATRIX_A}

	if mti >= _N {

		var kk int

		if mti == _N + 1 {
			Seed(uint32(5489))
		}

		for kk = 0; kk < _N - _M ; kk++ {
			y = (mt[kk] & _UPPER_MASK) | (mt[kk + 1] & _LOWER_MASK)
			mt[kk] = mt[kk + _M] ^ (y >> 1) ^ mag01[y & 1]
		}
		for ; kk < _N - 1; kk++ {
			y = (mt[kk] & _UPPER_MASK) | (mt[kk + 1] & _LOWER_MASK)
			mt[kk] = mt[kk + (_M - _N)] ^ (y >> 1) ^ mag01[y & 1]
		}
		y = (mt[_N - 1] & _UPPER_MASK) | (mt[0] & _LOWER_MASK)
		mt[_N - 1] = mt[_M - 1] ^ (y >> 1) ^ mag01[y & 1]

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

// The C source does this * 1.0 / foo, is the multiplication needed??
// It does generate different results for real1()

func genrand_real1() float64 {									// [0,1]
	return float64(genrand_int32()) * (1.0 / 4294967295.0)
}

func genrand_real2() float64 {									// [0,1)
	return float64(genrand_int32()) * (1.0 / 4294967296.0)
}

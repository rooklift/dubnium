package mt19937_32

import (
	"fmt"
	"testing"
)

func TestUint32(t *testing.T) {

	var z uint32

	for n := 0; n < 10000; n++ {
		z = Uint32()
	}

	if z != 4123659995 {
		t.Errorf("Expected 4123659995 after 10000 calls on unseeded RNG, got %v\n", z)
	}
}

func TestUrd(t *testing.T) {

	Seed(0)

	/*
	#include <random>
	#include <stdio.h>

	using namespace std;

	int main() {
		mt19937 rng(0);
		std::uniform_real_distribution<double> urd(0.0, 1.0);
		for (int n = 0; n < 10; n++) {
			printf("%.17f\n", urd(rng));
		}
	}
	*/

	cpp_output := []string{
		"0.59284461651668263",
		"0.84426574425659828",
		"0.85794561998982988",
		"0.84725173738433124",
		"0.62356369649610832",
		"0.38438170837375663",
		"0.29753460535723419",
		"0.05671297593316366",
		"0.27265629474158931",
		"0.47766511174464632",
	}

	for n := 0; n < 10; n++ {

		my_output := Urd()

		printed_to_17_dp := fmt.Sprintf("%.17f", my_output)

		if printed_to_17_dp != cpp_output[n] {
			t.Errorf("Expected %v from Urd(), got %v\n", cpp_output[n], printed_to_17_dp)
		}
	}
}

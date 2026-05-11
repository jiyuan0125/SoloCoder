package hashtable

func isPrime(n int) bool {
	if n <= 1 {
		return false
	}
	if n <= 3 {
		return true
	}
	if n%2 == 0 || n%3 == 0 {
		return false
	}
	
	for i := 5; i*i <= n; i += 6 {
		if n%i == 0 || n%(i+2) == 0 {
			return false
		}
	}
	return true
}

func NextPrime(n int) int {
	if n <= 1 {
		return 2
	}
	if n == 2 {
		return 2
	}
	if n%2 == 0 {
		n++
	}
	
	for !isPrime(n) {
		n += 2
	}
	return n
}

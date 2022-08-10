package version

func DetectMinorVersion() (int, error) {
	// return 2 for 0.2.x or 3 for 0.3.x
	return 0, nil
}

func DBbasedUserData() (bool, error) {
	// return true if user data is stored in database (0.3.3-0.3.5), false if config file (0.3.0-0.3.2)
	return false, nil
}

package version

// Get version of CasaOS by calling API.
func GetVersionFromAPI() (string, error) {
	// TODO
	return "", nil
}

// Detect minor version of CasaOS. It returns 2 for "0.2.x" or 3 for "0.3.x"
//
// (This is often useful when failing to get version from API because CasaOS is not running.)
func DetectMinorVersion() (int, error) {
	// TODO
	return 0, nil
}

// Check if user data is stored in database (true) or in config file (false)
//
// (user data is stored in config file for 0.3.0-0.3.2)
func IsUserDataInDatabase() (bool, error) {
	// TODO
	return false, nil
}

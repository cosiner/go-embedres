package embedres

import "io/ioutil"

func ReadFile(fs Fs, path string) ([]byte, error) {
	fd, err := fs.Open(path)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	return ioutil.ReadAll(fd)
}

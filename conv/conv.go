package conv

import (
	"strconv"
)

func ParseInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

func ParseInt(s string) (int, error) {
	i, err := strconv.ParseInt(s, 10, 64)
	return int(i), err
}

func ParseInt8(s string) (int8, error) {
	i, err := strconv.ParseInt(s, 10, 64)
	return int8(i), err
}

func ParseInts64(s []string) ([]int64, error) {
	nums := []int64{}
	for _, a := range s {
		n, err := ParseInt64(a)
		if err != nil {
			return nil, err
		}
		nums = append(nums, n)
	}
	return nums, nil
}

func ParseInts(s []string) ([]int, error) {
	nums := []int{}
	for _, a := range s {
		n, err := ParseInt(a)
		if err != nil {
			return nil, err
		}
		nums = append(nums, n)
	}
	return nums, nil
}

func ParseInts8(s []string) ([]int8, error) {
	nums := []int8{}
	for _, a := range s {
		n, err := ParseInt8(a)
		if err != nil {
			return nil, err
		}
		nums = append(nums, n)
	}
	return nums, nil
}

func FormatInt64(i int64) string {
	return strconv.FormatInt(i, 10)
}

func FormatInt(i int) string {
	return strconv.FormatInt(int64(i), 10)
}

func FormatInt8(i int8) string {
	return strconv.FormatInt(int64(i), 10)
}

func ParseBool(s string) (bool, error) {
	return strconv.ParseBool(s)
}

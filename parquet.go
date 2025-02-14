// Copyright 2022 Twilio Inc.

// Package parquet is a library for working with parquet files. For an overview
// of Parquet's qualities as a storage format, see this blog post:
// https://blog.twitter.com/engineering/en_us/a/2013/dremel-made-simple-with-parquet
//
// Or see the Parquet documentation: https://parquet.apache.org/documentation/latest/
package parquet

func atLeastOne(size int) int {
	return atLeast(size, 1)
}

func atLeast(size, least int) int {
	if size < least {
		return least
	}
	return size
}

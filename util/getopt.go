/*
Copyright 2019 Drew DeVault <sir@cmpwn.com>

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

1. Redistributions of source code must retain the above copyright notice, this
list of conditions and the following disclaimer.

2. Redistributions in binary form must reproduce the above copyright notice,
this list of conditions and the following disclaimer in the documentation
and/or other materials provided with the distribution.

3. Neither the name of the copyright holder nor the names of its contributors
may be used to endorse or promote products derived from this software without
specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/

package util

import (
	"fmt"
	"os"
)

// In the case of "-o example", Option is 'o' and "example" is Value. For
// options which do not take an argument, Value is "".
type Option struct {
	Option rune
	Value  string
}

// This is returned when an unknown option is found in argv, but not in the
// option spec.
type UnknownOptionError rune

func (e UnknownOptionError) Error() string {
	return fmt.Sprintf("%s: unknown option -%c", os.Args[0], rune(e))
}

// This is returned when an option with a mandatory argument is missing that
// argument.
type MissingOptionError rune

func (e MissingOptionError) Error() string {
	return fmt.Sprintf("%s: expected argument for -%c", os.Args[0], rune(e))
}

// Getopts implements a POSIX-compatible options interface.
//
// Returns a slice of options and the index of the first non-option argument.
//
// If an error is returned, you must print it to stderr to be POSIX complaint.
func Getopts(argv []string, spec string) ([]Option, int, error) {
	optmap := make(map[rune]bool)
	runes := []rune(spec)
	for i, rn := range spec {
		if rn == ':' {
			if i == 0 {
				continue
			}
			optmap[runes[i-1]] = true
		} else {
			optmap[rn] = false
		}
	}

	var (
		i    int
		opts []Option
	)
	for i = 1; i < len(argv); i++ {
		arg := argv[i]
		runes = []rune(arg)
		if len(arg) == 0 || arg == "-" {
			break
		}
		if arg[0] != '-' {
			break
		}
		if arg == "--" {
			i++
			break
		}
		for j, opt := range runes[1:] {
			if optopt, ok := optmap[opt]; !ok {
				opts = append(opts, Option{'?', ""})
				return opts, i, UnknownOptionError(opt)
			} else if optopt {
				if j+1 < len(runes)-1 {
					opts = append(opts, Option{opt, string(runes[j+2:])})
					break
				} else {
					if i+1 >= len(argv) {
						if len(spec) >= 1 && spec[0] == ':' {
							opts = append(opts, Option{':', string(opt)})
						} else {
							return opts, i, MissingOptionError(opt)
						}
					} else {
						opts = append(opts, Option{opt, argv[i+1]})
						i++
					}
				}
			} else {
				opts = append(opts, Option{opt, ""})
			}
		}
	}
	return opts, i, nil
}

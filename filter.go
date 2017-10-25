package main

/*
**   Copyright 2017 Telenor Digital AS
**
**  Licensed under the Apache License, Version 2.0 (the "License");
**  you may not use this file except in compliance with the License.
**  You may obtain a copy of the License at
**
**      http://www.apache.org/licenses/LICENSE-2.0
**
**  Unless required by applicable law or agreed to in writing, software
**  distributed under the License is distributed on an "AS IS" BASIS,
**  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
**  See the License for the specific language governing permissions and
**  limitations under the License.
 */
import "regexp"

// Filter is documented
type Filter struct {
	included []*regexp.Regexp
}

// NewFilter is documented
func NewFilter(regexps ...string) (*Filter, error) {
	ret := Filter{make([]*regexp.Regexp, 0)}
	for _, re := range regexps {
		compiled, err := regexp.Compile(re)
		if err != nil {
			return nil, err
		}
		ret.included = append(ret.included, compiled)
	}
	return &ret, nil
}

// Include is documented
func (f *Filter) Include(name string) bool {
	for _, v := range f.included {
		if v.Match([]byte(name)) {
			return true
		}
	}
	return false
}

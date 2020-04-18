// Copyright 2018 - 2020 Fabian Wenzelmann
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gummibaum

// IntMin returns the minimum of a and b.
func IntMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// MergeStringMaps combines two string maps. The result is a new map (both maps are
// unchanged) containing all entries from m1 and m2. If a key is present in both maps
// the value from m2 is used.
func MergeStringMaps(m1, m2 map[string]string) map[string]string {
	res := make(map[string]string, len(m1)+len(m2))
	for key, value := range m1 {
		res[key] = value
	}
	for key, value := range m2 {
		res[key] = value
	}
	return res
}

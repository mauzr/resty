/*
Copyright 2019 Alexander Sowitzki.

GNU Affero General Public License version 3 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://opensource.org/licenses/AGPL-3.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package args

import (
	"fmt"
	"net/url"
	"strconv"
	"time"
)

// Argument is an argument that can be collected from an URL.
type Argument func(*url.URL) error

// Collect all the given arguments.
func Collect(url *url.URL, arguments []Argument) error {
	for _, argument := range arguments {
		if err := argument(url); err != nil {
			return err
		}
	}
	return nil
}

func stringFromURL(url *url.URL, name string, optional bool) (string, bool, error) {
	values, ok := url.Query()[name]

	if !ok && optional {
		return "", false, nil
	}

	if !ok {
		return "", false, fmt.Errorf("URL does not contain argument %v", name)
	}

	if len(values) > 1 {
		return "", false, fmt.Errorf("URL contains %v multiple times", name)
	}

	if len(values) < 1 && !optional {
		return "", false, fmt.Errorf("URL does not contain %v", name)
	}

	if len(values) < 1 {
		return "", false, nil
	}

	return values[0], true, nil
}

// Bool represents a bool type argument.
func Bool(target *bool, name string, optional bool) Argument {
	return func(url *url.URL) error {
		rawValue, set, err := stringFromURL(url, name, optional)
		if !set || err != nil {
			return err
		}

		value, err := strconv.ParseBool(rawValue)
		if err != nil {
			return fmt.Errorf("Could not parse bool from URL argument %v: %v", name, err)
		}
		*target = value
		return nil
	}
}

// Stripped URL without query is returned.
func Stripped(target *string) Argument {
	return func(url *url.URL) error {
		*target = url.Scheme + url.Host + url.Path
		return nil
	}
}

// Int represents a integer argument.
func Int(target *int, name string, optional bool) Argument {
	return func(url *url.URL) error {
		rawValue, set, err := stringFromURL(url, name, optional)
		if !set || err != nil {
			return err
		}

		value, err := strconv.ParseInt(rawValue, 10, 0)
		if err != nil {
			return fmt.Errorf("Could not parse int from URL argument %v: %v", name, err)
		}
		*target = int(value)
		return nil
	}
}

// Float represents a float64 argument.
func Float(target *float64, name string, optional bool) Argument {
	return func(url *url.URL) error {
		rawValue, set, err := stringFromURL(url, name, optional)
		if !set || err != nil {
			return err
		}

		value, err := strconv.ParseFloat(rawValue, 64)
		if err != nil {
			return fmt.Errorf("Could not parse float from URL argument %v: %v", name, err)
		}
		*target = value
		return nil
	}
}

// Uint represents an unsigned integer argument.
func Uint(target *uint, name string, optional bool) Argument {
	return func(url *url.URL) error {
		rawValue, set, err := stringFromURL(url, name, optional)
		if !set || err != nil {
			return err
		}

		value, err := strconv.ParseUint(rawValue, 10, 0)
		if err != nil {
			return fmt.Errorf("Could not parse int from URL argument %v: %v", name, err)
		}
		*target = uint(value)
		return nil
	}
}

// Duration represents an time duration argument.
func Duration(target *time.Duration, name string, optional bool) Argument {
	return func(url *url.URL) error {
		rawValue, set, err := stringFromURL(url, name, optional)
		if !set || err != nil {
			return err
		}

		value, err := time.ParseDuration(rawValue)
		if err != nil {
			return fmt.Errorf("Could not parse duration from URL argument %v: %v", name, err)
		}
		*target = value
		return nil
	}
}

// String represents a string argument.
func String(target *string, name string, optional bool) Argument {
	return func(url *url.URL) error {
		value, set, err := stringFromURL(url, name, optional)
		if !set || err != nil {
			return err
		}
		*target = value
		return nil
	}
}

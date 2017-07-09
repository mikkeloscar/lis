# lis - backlight daemon
[![Travis BuildStatus](https://travis-ci.org/mikkeloscar/lis.svg?branch=master)](https://travis-ci.org/mikkeloscar/lis)

`lis` is a daemon to automatically dim/undim the screen backlight when a user
is idling/active, thus conserving energy and improving battery life.

### Dependencies

* libx11
* libxss
* systemd >= 183

## lisc


#### Commands

```
lisc set 50%
lisc set -5%
lisc set +5%

lisc status

lisc dpms off
lisc dpms on
```

#### Protocol

```
SET 50%
SET -5%
SET +5%
STATUS
DPMS OFF
DPMS ON

Response:

OK (optional msg)
ERROR err msg
```

## LICENSE

Copyright (C) 2016  Mikkel Oscar Lyderik Larsen

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.

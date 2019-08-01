OpenSimDash
===========
There didn't seem to be many F/OSS libraries for collecting telemetry data from racing games, such as Codemasters' Dirt Rally and F1 2018, so I decided to create one.

What it do?
===========
OpenSimDash reads telemetry data sent over UDP from games, and pushes it to connected HID devices and web clients to power DIY sim racing dashboards/indicators.

Supported Games
===============
* Dirt Rally 1.0 and 2.0
* More soon . . .

Why golang?
===========
Golang offers much of the performance of C, while providing many features of modern languages, and can still utilize native C libraries (though losing some safety features in the process).

Currently [karalabe/hid](github.com/karalabe/hid) is used for USB HID communication instead of [gousb](https://github.com/google/gousb) (or a custom C library wrapping winsock).  The latter wraps [libusb](https://github.com/libusb/libusb) which is a bit painful to get started with as compared to the former being self-contained and bundling [hidapi](https://github.com/signal11/hidapi).  I hope to add support for more USB HID devices like the [SLI-M](http://www.leobodnar.com/products/SLI-M/), [SLI-Pro](https://www.leobodnar.com/products/SLI-PRO/), as well as USB serial devices (like Arduino), which will require replacing the existing USB library/code anyway.

Alternatives
============

#### F/OSS
* https://github.com/Billiam/pygauge
* https://github.com/rafaelreinert/F1

#### Closed Source
* https://www.simhubdash.com
* https://fanaleds.com
* https://www.stryder-it.de/simdashboard/
* https://x-sim.de/software.php
* https://simracingapps.com
* Many random things on https://www.racedepartment.com


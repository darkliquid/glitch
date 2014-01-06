glitch
======

An image glitcher in golang.

Initially this is a more-or-less direct port of http://www.airtightinteractive.com/demos/js/imageglitcher/ while I'm learning Go, but when I'm more confident the glitching algorithms will be rewritten.

All the original javascript algorithms on which the initial build of this project is based were created by Felix Turner.

    Usage: glitch [-gbls] input_image output_image
      -b=5: Defines how much brightening to do (0-100) - shorthand syntax
      -brightness=5: Defines how much brightening to do (0-100)
      -g=5: Defines how much glitching to do (0-100) - shorthand syntax
      -glitch=5: Defines how much glitching to do (0-100)
      -l=true: Apply the scan line filter - shorthand syntax
      -s="my.host.name": Seed for the randomiser - shorthand syntax
      -scanlines=true: Apply the scan line filter
      -seed="my.host.name": Seed for the randomiser

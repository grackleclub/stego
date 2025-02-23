# cryptogif

## what?

A way to encode very small amounts of text with incredible ineffiency.

## why?

Just because.

## how?

Up to 256 colors are available in each frame's color palette. The darkest 16 colors (as determined by the sum of red, green, and blue values), are reserved for encoding data. Any pixel which would otherwise have one of the darkest 16 colors is assigned to the 17th darkest color.

#!/bin/sh

find .. -name \*.png | sort | while read -r file; do
	# Exceptions.
	case "$file" in
		# Editing only.
		*/src/*) continue ;;
		../assets/tiles/warpzone_*.png) continue ;;
		# Intentionally violating.
		../assets/sprites/clock_*.png) continue ;;
	esac
	f=$(
		convert \
			\( "$file" -depth 8 -alpha off \) \
			\( "$file" -depth 8 -alpha off +dither -remap cga_palette.pnm \) \
			-channel RGB \
			-metric RMSE -format '%[distortion]\n' -compare \
			INFO:
	)
	if [ "$f" !=  0 ]; then
		echo "convert \( '$file' -depth 8 -alpha off +dither -remap cga_palette.pnm \) \( '$file' -depth 8 -alpha extract \) -compose CopyOpacity -composite "$file"  # $f"
	fi
done

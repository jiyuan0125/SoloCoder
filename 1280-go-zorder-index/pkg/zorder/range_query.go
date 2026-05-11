package zorder

func QueryRanges(rect Rect2D) []Range {
	ranges := make([]Range, 0, 32)
	queryRangesRecursive(rect, 31, 0, 0, &ranges)
	return ranges
}

func queryRangesRecursive(rect Rect2D, bit int, xPrefix, yPrefix uint32, ranges *[]Range) {
	xMin := rect.Min.X
	xMax := rect.Max.X
	yMin := rect.Min.Y
	yMax := rect.Max.Y

	xStart := xPrefix
	xEnd := xPrefix + (1 << (bit + 1)) - 1
	yStart := yPrefix
	yEnd := yPrefix + (1 << (bit + 1)) - 1

	if xEnd < xMin || xStart > xMax || yEnd < yMin || yStart > yMax {
		return
	}

	if xStart >= xMin && xEnd <= xMax && yStart >= yMin && yEnd <= yMax {
		startCode := Encode2DBit(Point2D{X: xStart, Y: yStart})
		endCode := Encode2DBit(Point2D{X: xEnd, Y: yEnd})
		*ranges = append(*ranges, Range{Start: startCode, End: endCode})
		return
	}

	if bit < 0 {
		startCode := Encode2DBit(Point2D{X: xPrefix, Y: yPrefix})
		*ranges = append(*ranges, Range{Start: startCode, End: startCode})
		return
	}

	mask := uint32(1 << bit)
	queryRangesRecursive(rect, bit-1, xPrefix, yPrefix, ranges)
	queryRangesRecursive(rect, bit-1, xPrefix, yPrefix|mask, ranges)
	queryRangesRecursive(rect, bit-1, xPrefix|mask, yPrefix, ranges)
	queryRangesRecursive(rect, bit-1, xPrefix|mask, yPrefix|mask, ranges)
}

func QueryRangesOptimized(rect Rect2D) []Range {
	ranges := make([]Range, 0, 16)
	stack := make([]rangeStackItem, 0, 64)
	stack = append(stack, rangeStackItem{bit: 31, xPrefix: 0, yPrefix: 0})

	for len(stack) > 0 {
		item := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		xMin := rect.Min.X
		xMax := rect.Max.X
		yMin := rect.Min.Y
		yMax := rect.Max.Y

		xStart := item.xPrefix
		xEnd := item.xPrefix + (1 << (item.bit + 1)) - 1
		yStart := item.yPrefix
		yEnd := item.yPrefix + (1 << (item.bit + 1)) - 1

		if xEnd < xMin || xStart > xMax || yEnd < yMin || yStart > yMax {
			continue
		}

		if xStart >= xMin && xEnd <= xMax && yStart >= yMin && yEnd <= yMax {
			startCode := Encode2DBit(Point2D{X: xStart, Y: yStart})
			endCode := Encode2DBit(Point2D{X: xEnd, Y: yEnd})
			ranges = append(ranges, Range{Start: startCode, End: endCode})
			continue
		}

		if item.bit < 0 {
			startCode := Encode2DBit(Point2D{X: item.xPrefix, Y: item.yPrefix})
			ranges = append(ranges, Range{Start: startCode, End: startCode})
			continue
		}

		mask := uint32(1 << item.bit)
		stack = append(stack, rangeStackItem{bit: item.bit - 1, xPrefix: item.xPrefix | mask, yPrefix: item.yPrefix | mask})
		stack = append(stack, rangeStackItem{bit: item.bit - 1, xPrefix: item.xPrefix | mask, yPrefix: item.yPrefix})
		stack = append(stack, rangeStackItem{bit: item.bit - 1, xPrefix: item.xPrefix, yPrefix: item.yPrefix | mask})
		stack = append(stack, rangeStackItem{bit: item.bit - 1, xPrefix: item.xPrefix, yPrefix: item.yPrefix})
	}

	return ranges
}

type rangeStackItem struct {
	bit     int
	xPrefix uint32
	yPrefix uint32
}

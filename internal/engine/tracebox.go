package engine

import (
	m "github.com/divVerent/aaaaaa/internal/math"
)

func appendLineToTraces(traces map[m.Delta]struct{}, start, end m.Delta) {
	delta := end.Sub(start)
	length := delta.Norm1()
	traces[start] = struct{}{}
	for i := MinEntitySize; i < length; i += MinEntitySize {
		pos := start.Add(delta.Mul(i).Div(length))
		traces[pos] = struct{}{}
	}
	traces[end] = struct{}{}
}

// traceBox moves a size-sized box from from to to and yields info about where it hits solid etc.
func traceBox(w *World, from m.Rect, to m.Pos, o TraceOptions) TraceResult {
	// TODO make a real implementation.
	traces := map[m.Delta]struct{}{}
	delta := to.Delta(from.Origin)
	// TODO refactor using OppositeCorner?
	if delta.DX < 0 {
		appendLineToTraces(traces, m.Delta{DX: 0, DY: 0}, m.Delta{DX: 0, DY: from.Size.DY - 1})
	} else {
		appendLineToTraces(traces, m.Delta{DX: from.Size.DX - 1, DY: 0}, m.Delta{DX: from.Size.DX - 1, DY: from.Size.DY - 1})
	}
	if delta.DY < 0 {
		appendLineToTraces(traces, m.Delta{DX: 0, DY: 0}, m.Delta{DX: from.Size.DX - 1, DY: 0})
	} else {
		appendLineToTraces(traces, m.Delta{DX: 0, DY: from.Size.DY - 1}, m.Delta{DX: from.Size.DX - 1, DY: from.Size.DY - 1})
	}
	var result TraceResult
	var best int
	haveTrace := false
	for delta := range traces {
		trace := traceLine(w, from.Origin.Add(delta), to.Add(delta), o)
		adjustedEnd := trace.EndPos.Sub(delta)
		score := adjustedEnd.Delta(from.Origin).Norm1() * 2
		if trace.HitEntity == nil {
			// Get shortest trace, BUT prefer those that hit entities.
			score++
		}
		if !haveTrace || score < best {
			best = score
			haveTrace = true
			result.EndPos = adjustedEnd
			result.HitTilePos = trace.HitTilePos
			result.HitTile = trace.HitTile
			result.HitEntity = trace.HitEntity
			result.HitFogOfWar = trace.HitFogOfWar
		}
	}
	return result
}

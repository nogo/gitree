package graph

// Box-drawing characters for graph rendering
const (
	CharNode       = '●'
	CharNodeHollow = '○'
	CharVertical   = '│'
	CharHorizontal = '─'
	CharCornerTR   = '┐' // top-right: branch starts, goes down-left
	CharCornerBR   = '┘' // bottom-right: merge from right
	CharCornerTL   = '┌' // top-left: branch starts, goes down-right
	CharCornerBL   = '└' // bottom-left: merge from left
	CharTeeLeft    = '┤' // T-junction: line continues, branch left
	CharTeeRight   = '├' // T-junction: line continues, branch right
	CharTeeDown    = '┬' // T-junction: line continues, branches down
	CharTeeUp      = '┴' // T-junction: line continues, branches up
	CharCross      = '┼' // cross: two lines crossing
	CharSpace      = ' '
)

// CellType represents what kind of graph element occupies a cell
type CellType uint8

const (
	CellEmpty CellType = iota
	CellNode
	CellNodeHollow
	CellPassThrough  // vertical line continues
	CellHorizontal   // horizontal connection
	CellCornerTR     // ┐
	CellCornerBR     // ┘
	CellCornerTL     // ┌
	CellCornerBL     // └
	CellTeeRight     // ├
	CellTeeLeft      // ┤
	CellTeeDown      // ┬
	CellTeeUp        // ┴
	CellCross        // ┼
)

// CellChar returns the character for a cell type
func CellChar(ct CellType) rune {
	switch ct {
	case CellNode:
		return CharNode
	case CellNodeHollow:
		return CharNodeHollow
	case CellPassThrough:
		return CharVertical
	case CellHorizontal:
		return CharHorizontal
	case CellCornerTR:
		return CharCornerTR
	case CellCornerBR:
		return CharCornerBR
	case CellCornerTL:
		return CharCornerTL
	case CellCornerBL:
		return CharCornerBL
	case CellTeeRight:
		return CharTeeRight
	case CellTeeLeft:
		return CharTeeLeft
	case CellTeeDown:
		return CharTeeDown
	case CellTeeUp:
		return CharTeeUp
	case CellCross:
		return CharCross
	default:
		return CharSpace
	}
}

// SelectChar determines the character based on connection directions
// above/below = vertical connections, left/right = horizontal connections
func SelectChar(above, below, left, right, isNode bool) rune {
	if isNode {
		// Node with connections
		switch {
		case !above && !below && left && right:
			return CharNode // node on horizontal line (rare)
		case above && below && left && !right:
			return CharNode // could use special char, but ● is clear
		case above && below && !left && right:
			return CharNode
		default:
			return CharNode
		}
	}

	// Connection characters based on directions
	switch {
	case above && below && !left && !right:
		return CharVertical
	case !above && !below && left && right:
		return CharHorizontal
	case !above && below && !left && right:
		return CharCornerTL // ┌
	case !above && below && left && !right:
		return CharCornerTR // ┐
	case above && !below && !left && right:
		return CharCornerBL // └
	case above && !below && left && !right:
		return CharCornerBR // ┘
	case above && below && !left && right:
		return CharTeeRight // ├
	case above && below && left && !right:
		return CharTeeLeft // ┤
	case !above && below && left && right:
		return CharTeeDown // ┬
	case above && !below && left && right:
		return CharTeeUp // ┴
	case above && below && left && right:
		return CharCross // ┼
	default:
		return CharSpace
	}
}

package main

import "github.com/go-gl/gl/v4.1-core/gl"

type Cell struct {
	drawable uint32

	alive     bool
	aliveNext bool
	
	timeAlive int

	x int
	y int
}


var (
	square = []float32{
		-0.5, 0.5, 0,
		-0.5, -0.5, 0,
		0.5, -0.5, 0,

		-0.5, 0.5, 0,
		0.5, 0.5, 0,
		0.5, -0.5, 0,
	}
)

func (c *Cell) Draw() {
	if !c.alive {
		return
	}

	gl.BindVertexArray(c.drawable)
	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(square)/3))
}

func (c *Cell) Update(cells [][]*Cell) {
	c.checkState(cells)
	if c.alive {
		c.timeAlive = c.timeAlive + 1
	} else {
		c.timeAlive = 0
	}
}

func (c *Cell) checkState(cells [][]*Cell) {
	c.alive = c.aliveNext
	c.aliveNext = c.alive

	liveCount := c.liveNeighbors(cells)
	if c.alive {
		// cells with fewer than two live neighbors dies
		if liveCount < 2 {
			c.aliveNext = false
		}

		// A live Cell with 2 or 3 live neighbors lives
		if liveCount == 2 || liveCount == 3 {
			c.aliveNext = true
		}

		// A live Cell with more than 3 neighbors dies
		if liveCount > 3 {
			c.aliveNext = false
		}
	} else {
		if liveCount == 3 {
			c.aliveNext = true
		}
	}
}

func (c *Cell) liveNeighbors(cells [][]*Cell) int {
	var liveCount int
	add := func(x, y int) {
		// If we're at an edge, check the other side of the board.
		if x == len(cells) {
			x = 0
		} else if x == -1 {
			x = len(cells) - 1
		}
		if y == len(cells[x]) {
			y = 0
		} else if y == -1 {
			y = len(cells[x]) - 1
		}

		if cells[x][y].alive {
			liveCount++
		}
	}

	add(c.x-1, c.y)   // To the left
	add(c.x+1, c.y)   // To the right
	add(c.x, c.y+1)   // up
	add(c.x, c.y-1)   // down
	add(c.x-1, c.y+1) // top-left
	add(c.x+1, c.y+1) // top-right
	add(c.x-1, c.y-1) // bottom-left
	add(c.x+1, c.y-1) // bottom-right

	return liveCount
}

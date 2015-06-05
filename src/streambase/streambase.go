//==================================================================================
// Author: Stoney Kang, sikang99@gmail.com, 2015
// Base information for streaming
// Author : Stoney Kang, sikang99@gmail.com
//==================================================================================

package streambase

import "errors"

//----------------------------------------------------------------------------------
const (
	KBYTE = 1024
	MBYTE = 1024 * KBYTE // Kilo
	GBYTE = 1024 * MBYTE // Giga
	TBYTE = 1024 * GBYTE // Tera
	HBYTE = 1024 * TBYTE // Hexa
)

const (
	STATUS_IDLE = iota
	STATUS_USING
)

var (
	ErrEmpty  = errors.New("empty")
	ErrFull   = errors.New("full")
	ErrStatus = errors.New("invalid status")
)

// ---------------------------------E-----N-----D-----------------------------------

package hemfix

/*

goupx: Fix compiled go binaries so that they can be packed by
       the universal packer for executables (upx)

Copyright (c) 2012 Peter Waller <peter@pwaller.net>
All rights reserved.

Based on code found at http://sourceforge.net/tracker/?func=detail&atid=102331&aid=3408066&group_id=2331

Based on hemfix.c Copyright (C) 2012 John Reiser, BitWagon Software LLC

  This program is free software; you can redistribute it and/or modify
  it under the terms of the GNU General Public License as published by
  the Free Software Foundation; either version 3, or (at your option)
  any later version.

  This program is distributed in the hope that it will be useful,
  but WITHOUT ANY WARRANTY; without even the implied warranty of
  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
  GNU General Public License for more details.

  You should have received a copy of the GNU General Public License
  along with this program; if not, write to the Free Software Foundation,
  Inc., 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301, USA.
*/

import (
	ELF "debug/elf"
	"encoding/binary"
	"errors"
	"io"
	"log"
	"os"
)

// The functions gethdr and writephdr are heavily influenced by code found at
// http://golang.org/src/pkg/debug/elf/file.go

// Returns the Prog header offset and size
func gethdr(f *ELF.File, sr io.ReadSeeker) (int64, int, error) {
	sr.Seek(0, os.SEEK_SET)

	switch f.Class {
	case ELF.ELFCLASS32:
		hdr := new(ELF.Header32)
		if err := binary.Read(sr, f.ByteOrder, hdr); err != nil {
			return 0, 0, err
		}
		return int64(hdr.Phoff), int(hdr.Phentsize), nil

	case ELF.ELFCLASS64:
		hdr := new(ELF.Header64)
		if err := binary.Read(sr, f.ByteOrder, hdr); err != nil {
			return 0, 0, err
		}
		return int64(hdr.Phoff), int(hdr.Phentsize), nil
	}
	return 0, 0, errors.New("Unexpected ELF class")
}

// Write out a Prog header to an elf with a given destination
// Writes out `p` to `sw` at `dst` using information from `f`
func writephdr(f *ELF.File, dst int64, sw io.WriteSeeker, p *ELF.Prog) error {
	sw.Seek(dst, os.SEEK_SET)

	switch f.Class {
	case ELF.ELFCLASS32:
		hdr := ELF.Prog32{
			Type:   uint32(p.Type),
			Flags:  uint32(p.Flags),
			Off:    uint32(p.Off),
			Vaddr:  uint32(p.Vaddr),
			Paddr:  uint32(p.Paddr),
			Filesz: uint32(p.Filesz),
			Memsz:  uint32(p.Memsz),
			Align:  uint32(p.Align),
		}
		if err := binary.Write(sw, f.ByteOrder, hdr); err != nil {
			return err
		}

	case ELF.ELFCLASS64:
		hdr := ELF.Prog64{
			Type:   uint32(p.Type),
			Flags:  uint32(p.Flags),
			Off:    p.Off,
			Vaddr:  p.Vaddr,
			Paddr:  p.Paddr,
			Filesz: p.Filesz,
			Memsz:  p.Memsz,
			Align:  p.Align,
		}
		if err := binary.Write(sw, f.ByteOrder, hdr); err != nil {
			return err
		}
	}
	return nil
}

func fixelf(elf *ELF.File, fd io.ReadWriteSeeker) error {

	// Determine where to write header (need
	off, sz, err := gethdr(elf, fd)
	if err != nil {
		return err
	}

	for i := range elf.Progs {
		p := elf.Progs[i]

		if p.ProgHeader.Type != ELF.PT_LOAD {
			// Only consider PT_LOAD sections
			continue
		}

		if p.Flags&ELF.PF_X != ELF.PF_X {
			continue
		}

		mask := -p.Align
		if ^mask&p.Vaddr != 0 && (^mask&(p.Vaddr-p.Off)) == 0 {
			log.Printf("Hemming PT_LOAD section")
			hem := ^mask & p.Off
			p.Off -= hem
			p.Vaddr -= hem
			if p.Paddr != 0 {
				p.Paddr -= hem
			}
			p.Filesz += hem
			p.Memsz += hem

			dst := off + int64(sz*i)
			writephdr(elf, dst, fd, p)
			break
		}
	}
	return nil
}

func FixFile(filename string) error {
	fd, err := os.OpenFile(filename, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	defer fd.Close()

	elf, err := ELF.NewFile(fd)
	if err != nil {
		log.Print("Failed to parse ELF. This can happen if the binary is already packed.")
		return err
	}
	defer elf.Close()

	log.Printf("%+v", elf.FileHeader)
	err = fixelf(elf, fd)
	if err != nil {
		log.Fatal("Failure to read ELF header")
		return err
	}
	return nil
}

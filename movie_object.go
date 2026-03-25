package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

type MovieObjectBDMV struct {
	Version string
	Objects []MovieObject
}

type MovieObject struct {
	ResumeIntentionFlag bool
	MenuCallMask        bool
	TitleSearchMask     bool
	Commands            []NavigationCommand
}

type NavigationCommand struct {
	Instruction uint32
	Destination uint32
	Source      uint32
}

func ParseMovieObject(path string) (*MovieObjectBDMV, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	typeInd := make([]byte, 4)
	version := make([]byte, 4)
	io.ReadFull(f, typeInd)
	io.ReadFull(f, version)

	if string(typeInd) != typeIndicatorMOBJ {
		return nil, fmt.Errorf("not a MovieObject.bdmv file (got %q)", typeInd)
	}

	mobj := &MovieObjectBDMV{Version: string(version)}

	f.Seek(movieObjectTableOffset, io.SeekStart)

	var tableLen uint32
	binary.Read(f, binary.BigEndian, &tableLen)

	var reserved uint32
	binary.Read(f, binary.BigEndian, &reserved)

	var numObjects uint16
	binary.Read(f, binary.BigEndian, &numObjects)

	mobj.Objects = nil
	for {
		obj, err := readMovieObject(f)
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("reading object %d: %w", len(mobj.Objects), err)
		}
		mobj.Objects = append(mobj.Objects, obj)
	}

	return mobj, nil
}

func readMovieObject(r io.Reader) (MovieObject, error) {
	var flags uint16
	if err := binary.Read(r, binary.BigEndian, &flags); err != nil {
		return MovieObject{}, err
	}

	var numCmds uint16
	if err := binary.Read(r, binary.BigEndian, &numCmds); err != nil {
		return MovieObject{}, err
	}

	obj := MovieObject{
		ResumeIntentionFlag: flags&mobjFlagResumeIntention != 0,
		MenuCallMask:        flags&mobjFlagMenuCallMask != 0,
		TitleSearchMask:     flags&mobjFlagTitleSearchMask != 0,
		Commands:            make([]NavigationCommand, numCmds),
	}

	for i := range obj.Commands {
		cmd, err := readNavCommand(r)
		if err != nil {
			return MovieObject{}, fmt.Errorf("reading command %d: %w", i, err)
		}
		obj.Commands[i] = cmd
	}

	return obj, nil
}

func readNavCommand(r io.Reader) (NavigationCommand, error) {
	var cmd NavigationCommand
	if err := binary.Read(r, binary.BigEndian, &cmd.Instruction); err != nil {
		return cmd, err
	}
	if err := binary.Read(r, binary.BigEndian, &cmd.Destination); err != nil {
		return cmd, err
	}
	if err := binary.Read(r, binary.BigEndian, &cmd.Source); err != nil {
		return cmd, err
	}
	return cmd, nil
}

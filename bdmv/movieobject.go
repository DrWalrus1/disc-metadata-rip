package bdmv

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

// MovieObjectBDMV represents the parsed contents of a MovieObject.bdmv file.
type MovieObjectBDMV struct {
	Version string
	Objects []MovieObject
}

// MovieObject represents a single navigation object.
type MovieObject struct {
	ResumeIntentionFlag bool
	MenuCallMask        bool
	TitleSearchMask     bool
	Commands            []NavigationCommand
}

// NavigationCommand represents a single HDMV navigation command.
type NavigationCommand struct {
	Instruction uint32
	Destination uint32
	Source      uint32
}

// ParseMovieObject parses the MovieObject.bdmv file at the given path.
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

	if string(typeInd) != TypeIndicatorMOBJ {
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

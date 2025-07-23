package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type Formats int

const (
	JSON Formats = iota
	Pretty
	Raw
)

type CLIWriter struct {
	writer io.Writer
	format Formats
}

func New(outputPath string, format Formats) (*CLIWriter, error) {
	var w io.Writer

	if outputPath != "" {
		file, err := os.Create(outputPath)
		if err != nil {
			return nil, err
		}
		w = file
	} else {
		w = os.Stdout
	}

	return &CLIWriter{
		writer: w,
		format: format,
	}, nil
}

func (o *CLIWriter) Write(data any) error {
	switch o.format {
	case JSON:
		return json.NewEncoder(o.writer).Encode(data)
	case Pretty:
		enc := json.NewEncoder(o.writer)
		enc.SetIndent("", "  ")
		return enc.Encode(data)
	case Raw:
		d, ok := data.([]byte)
		if !ok {
			return fmt.Errorf("raw output format option requires []byte")
		}
		_, err := o.writer.Write(d)
		return err
	default:
		return fmt.Errorf("unknown format option provided")
	}
}

func (o *CLIWriter) Close() error {
	if o.writer == os.Stdout || o.writer == os.Stderr {
		return nil
	}

	if closer, ok := o.writer.(io.Closer); ok {
		if o.writer != os.Stdout {
			return closer.Close()
		}
		return nil
	}

	return fmt.Errorf("unable to close writer")
}

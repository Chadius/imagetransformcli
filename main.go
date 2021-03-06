package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/cserrant/image-transform-cli/command"
	"image"
	"image/png"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func main() {
	commandLineArguments := getCommandLineArguments()

	inputImageDataByteStream, _ := ioutil.ReadFile(commandLineArguments.SourceImageFilename)

	formulaDataByteStream, formulaErr := ioutil.ReadFile(commandLineArguments.FormulaFilename)
	if formulaErr != nil {
		log.Fatal(formulaErr)
	}

	outputSettingsJSONByteStream := []byte(
		fmt.Sprintf(
			"{'output_width':%d,'output_height':%d}",
			commandLineArguments.OutputWidth,
			commandLineArguments.OutputHeight,
		),
	)

	var output bytes.Buffer

	useServerURL := false
	if commandLineArguments.ServerURL != "" {
		useServerURL = true
	}

	commandProcessor := command.NewCommandProcessor(nil, nil)
	processingErr := commandProcessor.ProcessArgumentsToTransformImage(&command.TransformArguments{
		InputImageData:     inputImageDataByteStream,
		FormulaData:        formulaDataByteStream,
		OutputSettingsData: outputSettingsJSONByteStream,
		OutputImageData:    &output,
		ServerURL:          commandLineArguments.ServerURL,
		UseServerURL:       useServerURL,
	})

	if processingErr != nil {
		log.Fatal(processingErr)
	}

	outputReader := bytes.NewReader(output.Bytes())
	outputImage, _ := png.Decode(outputReader)
	outputToFile(commandLineArguments.OutputFilename, outputImage)
}

// CommandLineArguments assume the user provides filenames to create a pattern.
type CommandLineArguments struct {
	FormulaFilename     string
	OutputFilename      string
	OutputHeight        int
	OutputWidth         int
	SourceImageFilename string
	ServerURL           string
}

func getCommandLineArguments() *CommandLineArguments {
	var sourceImageFilename, outputFilename, outputDimensions, serverURL string
	formulaFilename := "data/oldformula.yml"
	flag.StringVar(&formulaFilename, "f", "data/oldformula.yml", "See -oldformula")
	flag.StringVar(&formulaFilename, "oldformula", "data/oldformula.yml", "The filename of the oldformula file. Defaults to data/oldformula.yml")

	flag.StringVar(&sourceImageFilename, "in", "", "See -source. Required.")
	flag.StringVar(&sourceImageFilename, "source", "", "Source filename. Required.")

	flag.StringVar(&outputFilename, "out", "", "Output filename. Required.")
	outputDimensions = "200x200"
	flag.StringVar(&outputDimensions, "size", "200x200", "Output size in pixels, separated with an x. Default to 200x200.")

	flag.StringVar(&serverURL, "url", "", "URL of the transform server. If blank, it assumes the transformer is installed on this machine.")
	serverURL = ""

	flag.Parse()

	checkSourceArgument(sourceImageFilename)
	outputWidth, outputHeight := checkOutputArgument(outputFilename, outputDimensions)

	return &CommandLineArguments{
		FormulaFilename:     formulaFilename,
		OutputFilename:      outputFilename,
		OutputHeight:        outputHeight,
		OutputWidth:         outputWidth,
		SourceImageFilename: sourceImageFilename,
		ServerURL:           serverURL,
	}
}

func checkSourceArgument(sourceImageFilename string) {
	if sourceImageFilename == "" {
		log.Fatal("missing source filename")
	}
}

func openSourceImage(filenameArguments *CommandLineArguments) image.Image {
	reader, err := os.Open(filenameArguments.SourceImageFilename)
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	colorSourceImage, _, err := image.Decode(reader)
	if err != nil {
		log.Fatal(err)
	}
	return colorSourceImage
}

func outputToFile(outputFilename string, outputImage image.Image) {
	err := os.MkdirAll(filepath.Dir(outputFilename), 0777)
	if err != nil {
		panic(err)
	}
	outputImageFile, err := os.Create(outputFilename)
	if err != nil {
		panic(err)
	}
	defer outputImageFile.Close()
	png.Encode(outputImageFile, outputImage)
}

func checkOutputArgument(outputFilename, outputDimensions string) (int, int) {
	if outputFilename == "" {
		log.Fatal("missing output filename")
	}

	outputWidth, widthErr := strconv.Atoi(strings.Split(outputDimensions, "x")[0])
	if widthErr != nil {
		log.Fatal(widthErr)
	}

	outputHeight, heightErr := strconv.Atoi(strings.Split(outputDimensions, "x")[1])
	if heightErr != nil {
		log.Fatal(heightErr)
	}

	return outputWidth, outputHeight
}

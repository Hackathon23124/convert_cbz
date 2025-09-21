package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

// ANSI color codes for professional logging
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorCyan   = "\033[36m"
)

// ConversionStats tracks overall conversion statistics
type ConversionStats struct {
	mu            sync.Mutex
	Total         int
	Success       int
	Errors        int
	Skipped       int
	NonImageFiles int
}

// WorkItem represents a single conversion job
type WorkItem struct {
	FolderName string
	SourcePath string
	OutputPath string
}

func main() {
	// Command line argument parsing
	var (
		inputDir    = flag.String("input", "", "Input directory containing folders to convert (required)")
		outputDir   = flag.String("output", "", "Output directory for CBZ files (required)")
		threads     = flag.Int("threads", 4, "Number of concurrent threads")
		showHelp    = flag.Bool("help", false, "Show usage information")
		showVersion = flag.Bool("version", false, "Show version information")
	)

	flag.Parse()

	// Handle version flag
	if *showVersion {
		fmt.Println("CBZ Converter v1.0.0")
		fmt.Println("Converts folders containing images to CBZ comic book archives")
		return
	}

	// Handle help flag or missing required arguments
	if *showHelp || *inputDir == "" || *outputDir == "" {
		showUsage()
		return
	}

	// Validate thread count - ensure reasonable bounds
	if *threads < 1 {
		*threads = 1
	} else if *threads > runtime.NumCPU()*2 {
		// Limit to 2x CPU cores to prevent resource exhaustion
		*threads = runtime.NumCPU() * 2
		logInfo(fmt.Sprintf("Thread count limited to %d (2x CPU cores)", *threads))
	}

	// Validate input directory exists
	if _, err := os.Stat(*inputDir); os.IsNotExist(err) {
		logError(fmt.Sprintf("Input directory does not exist: %s", *inputDir))
		os.Exit(1)
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		logError(fmt.Sprintf("Failed to create output directory: %v", err))
		os.Exit(1)
	}

	logInfo(fmt.Sprintf("Starting CBZ conversion with %d threads", *threads))
	logInfo(fmt.Sprintf("Input:  %s", *inputDir))
	logInfo(fmt.Sprintf("Output: %s", *outputDir))

	// Get list of folders to process
	folders, err := getFolders(*inputDir)
	if err != nil {
		logError(fmt.Sprintf("Failed to read input directory: %v", err))
		os.Exit(1)
	}

	if len(folders) == 0 {
		logWarning("No folders found in input directory")
		return
	}

	logInfo(fmt.Sprintf("Found %d folders to process", len(folders)))

	// Create work items
	workItems := make([]WorkItem, len(folders))
	for i, folder := range folders {
		workItems[i] = WorkItem{
			FolderName: folder,
			SourcePath: filepath.Join(*inputDir, folder),
			OutputPath: filepath.Join(*outputDir, folder+".cbz"),
		}
	}

	// Process folders concurrently
	stats := &ConversionStats{Total: len(folders)}
	processConcurrently(workItems, *threads, stats)

	// Print final statistics
	printFinalStats(stats)
}

func showUsage() {
	fmt.Println("CBZ Converter - Convert image folders to CBZ comic book archives")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Printf("  %s -input <folder> -output <folder> [options]\n", os.Args[0])
	fmt.Println()
	fmt.Println("REQUIRED:")
	fmt.Println("  -input   string    Input directory containing folders to convert")
	fmt.Println("  -output  string    Output directory for CBZ files")
	fmt.Println()
	fmt.Println("OPTIONS:")
	fmt.Println("  -threads int       Number of concurrent threads (default: 4)")
	fmt.Println("  -help             Show this help message")
	fmt.Println("  -version          Show version information")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Printf("  %s -input ./manga -output ./cbz\n", os.Args[0])
	fmt.Printf("  %s -input /home/user/comics -output /home/user/cbz -threads 8\n", os.Args[0])
	fmt.Println()
	fmt.Println("The program will:")
	fmt.Println("  1. Scan each folder in the input directory")
	fmt.Println("  2. Detect image files using MIME type analysis")
	fmt.Println("  3. Create compressed CBZ files in the output directory")
	fmt.Println("  4. Skip existing CBZ files to avoid overwriting")
	fmt.Println("  5. Report non-image files found but not included")
}

func getFolders(dir string) ([]string, error) {
	var folders []string

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			folders = append(folders, entry.Name())
		}
	}

	// Sort for consistent processing order
	sort.Strings(folders)
	return folders, nil
}

func processConcurrently(workItems []WorkItem, numThreads int, stats *ConversionStats) {
	// Create work channel with buffer to prevent blocking
	workChan := make(chan WorkItem, numThreads)

	// Create wait group to track completion
	var wg sync.WaitGroup

	// Start worker goroutines
	for i := range numThreads {
		wg.Add(1)
		go worker(i+1, workChan, &wg, stats)
	}

	// Send work items to channel
	go func() {
		defer close(workChan)
		for _, item := range workItems {
			workChan <- item
		}
	}()

	// Wait for all workers to complete
	wg.Wait()
}

func worker(id int, workChan <-chan WorkItem, wg *sync.WaitGroup, stats *ConversionStats) {
	defer wg.Done()

	for item := range workChan {
		// Process single conversion job
		processWorkItem(id, item, stats)

		// Small delay to prevent overwhelming the system
		time.Sleep(5 * time.Millisecond)
	}
}

func processWorkItem(workerID int, item WorkItem, stats *ConversionStats) {
	prefix := fmt.Sprintf("[WORKER %d]", workerID)

	logInfo(fmt.Sprintf("%s Processing: %s", prefix, truncateString(item.FolderName, 60)))

	// Check if output already exists
	if _, err := os.Stat(item.OutputPath); err == nil {
		logWarning(fmt.Sprintf("%s CBZ already exists, skipping: %s", prefix, filepath.Base(item.OutputPath)))
		stats.mu.Lock()
		stats.Skipped++
		stats.mu.Unlock()
		return
	}

	// Convert folder to CBZ
	nonImageCount, err := convertToCBZ(item.SourcePath, item.OutputPath)
	if err != nil {
		logError(fmt.Sprintf("%s Conversion failed: %v", prefix, err))
		stats.mu.Lock()
		stats.Errors++
		stats.mu.Unlock()
		return
	}

	// Update statistics
	stats.mu.Lock()
	stats.Success++
	stats.NonImageFiles += nonImageCount
	stats.mu.Unlock()

	logOK(fmt.Sprintf("%s Created: %s", prefix, filepath.Base(item.OutputPath)))

	// Report non-image files if found
	if nonImageCount > 0 {
		logWarning(fmt.Sprintf("%s Found %d non-image files (excluded from CBZ)", prefix, nonImageCount))
	}
}

func convertToCBZ(sourceDir, cbzPath string) (int, error) {
	// Scan directory for image and non-image files
	imageFiles, nonImageFiles, err := analyzeDirectory(sourceDir)
	if err != nil {
		return 0, fmt.Errorf("failed to analyze directory: %w", err)
	}

	if len(imageFiles) == 0 {
		return 0, fmt.Errorf("no image files found")
	}

	// Create CBZ file (which is just a ZIP with .cbz extension)
	cbzFile, err := os.Create(cbzPath)
	if err != nil {
		return 0, fmt.Errorf("failed to create CBZ file: %w", err)
	}
	defer cbzFile.Close()

	// Create ZIP writer with compression
	zipWriter := zip.NewWriter(cbzFile)
	defer zipWriter.Close()

	// Add all image files to the ZIP archive
	for _, imagePath := range imageFiles {
		if err := addFileToZip(zipWriter, imagePath, sourceDir); err != nil {
			return 0, fmt.Errorf("failed to add file to archive: %w", err)
		}
	}

	return len(nonImageFiles), nil
}

func analyzeDirectory(dir string) ([]string, []string, error) {
	var imageFiles []string
	var nonImageFiles []string

	// Walk through directory tree recursively
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories, only process files
		if d.IsDir() {
			return nil
		}

		// Determine if file is an image using MIME type detection
		isImage, err := isImageFile(path)
		if err != nil {
			// If MIME detection fails, include file anyway (fail-safe approach)
			// This prevents losing files due to permission issues or corrupted headers
			logWarning(fmt.Sprintf("Could not determine file type for %s, including in archive", filepath.Base(path)))
			imageFiles = append(imageFiles, path)
		} else if isImage {
			imageFiles = append(imageFiles, path)
		} else {
			// Track non-image files for reporting purposes
			nonImageFiles = append(nonImageFiles, filepath.Base(path))
		}

		return nil
	})

	if err != nil {
		return nil, nil, err
	}

	// Sort files for consistent ordering in the archive
	// This ensures pages appear in the correct sequence
	sort.Strings(imageFiles)
	sort.Strings(nonImageFiles)

	return imageFiles, nonImageFiles, nil
}

func isImageFile(filePath string) (bool, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return false, err
	}
	defer file.Close()

	// Read first 512 bytes for MIME type detection
	// This is sufficient for http.DetectContentType to identify most formats
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil && err != io.EOF {
		return false, err
	}

	// Use Go's built-in MIME type detection
	// This checks file headers/magic bytes rather than relying on extensions
	mimeType := http.DetectContentType(buffer)

	// Any MIME type starting with "image/" is considered an image
	// This includes JPEG, PNG, GIF, WebP, HEIF, AVIF, etc.
	return strings.HasPrefix(mimeType, "image/"), nil
}

func addFileToZip(zipWriter *zip.Writer, filePath, baseDir string) error {
	// Calculate relative path for the ZIP entry
	// This preserves the directory structure within the archive
	relPath, err := filepath.Rel(baseDir, filePath)
	if err != nil {
		return err
	}

	// Convert to forward slashes for ZIP standard compliance
	relPath = filepath.ToSlash(relPath)

	// Open source file
	sourceFile, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// Get file information for archive header
	fileInfo, err := sourceFile.Stat()
	if err != nil {
		return err
	}

	// Create ZIP file header
	header, err := zip.FileInfoHeader(fileInfo)
	if err != nil {
		return err
	}

	// Set compression method and file path
	header.Name = relPath
	header.Method = zip.Deflate // Use compression to reduce file size

	// Create ZIP entry
	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	// Copy file content to ZIP entry
	_, err = io.Copy(writer, sourceFile)
	return err
}

func printFinalStats(stats *ConversionStats) {
	stats.mu.Lock()
	defer stats.mu.Unlock()

	logInfo("Conversion completed")
	logInfo(fmt.Sprintf("Total folders:     %d", stats.Total))
	logOK(fmt.Sprintf("Successful:        %d", stats.Success))

	if stats.Skipped > 0 {
		logWarning(fmt.Sprintf("Skipped:           %d", stats.Skipped))
	}

	if stats.Errors > 0 {
		logError(fmt.Sprintf("Errors:            %d", stats.Errors))
	}

	if stats.NonImageFiles > 0 {
		logInfo(fmt.Sprintf("Non-image files:   %d (excluded)", stats.NonImageFiles))
	}

	// Calculate success rate
	processed := stats.Success + stats.Errors
	if processed > 0 {
		successRate := float64(stats.Success) / float64(processed) * 100
		logInfo(fmt.Sprintf("Success rate:      %.1f%%", successRate))
	}
}

// Logging functions with colored output and professional tags

func logInfo(message string) {
	fmt.Printf("%s[INFO]%s %s\n", ColorBlue, ColorReset, message)
}

func logOK(message string) {
	fmt.Printf("%s[OK]%s %s\n", ColorGreen, ColorReset, message)
}

func logWarning(message string) {
	fmt.Printf("%s[WARN]%s %s\n", ColorYellow, ColorReset, message)
}

func logError(message string) {
	fmt.Printf("%s[ERROR]%s %s\n", ColorRed, ColorReset, message)
}

func truncateString(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return string(runes[:maxLen])
	}
	// Use ellipsis to indicate truncation
	return string(runes[:maxLen-3]) + "..."
}

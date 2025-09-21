# CBZ Converter

A high-performance, concurrent tool for converting folders containing images into CBZ (Comic Book Archive) files. Built in Go for speed and reliability.

## Features

- **High Performance**: Multi-threaded processing with configurable concurrency
- **Smart Detection**: MIME-type based image detection (supports JPEG, PNG, GIF, WebP, HEIF, AVIF, and more)
- **Cross-Platform**: Runs on Linux, macOS, and various Unix systems
- **Professional Logging**: Color-coded output with detailed progress tracking
- **Safe Operations**: Skips existing files, handles errors gracefully
- **Comprehensive Reporting**: Detailed statistics and non-image file detection

## Installation

### Quick Build

For your current system:
```bash
./build.sh
```

For all supported architectures:
```bash
./build.sh all
```

The compiled binaries will be placed in the `./bin/` directory.

### Manual Build

If you prefer to build manually:
```bash
go build -o convert_cbz main.go
```

## Usage

### Basic Syntax
```bash
convert_cbz -input <input_folder> -output <output_folder> [options]
```

### Options

| Flag | Description | Default |
|------|-------------|---------|
| `-input` | Input directory containing folders to convert | *required* |
| `-output` | Output directory for CBZ files | *required* |
| `-threads` | Number of concurrent processing threads | `4` |
| `-help` | Show usage information | - |
| `-version` | Show version information | - |

### Examples

Convert manga folders to CBZ files:
```bash
convert_cbz -input ./manga -output ./cbz
```

Use 8 threads for faster processing:
```bash
convert_cbz -input /home/user/comics -output /home/user/cbz -threads 8
```

Show help:
```bash
convert_cbz -help
```

## How It Works

1. **Directory Scanning**: Recursively scans each folder in the input directory
2. **Image Detection**: Uses MIME type analysis to identify image files (not relying on file extensions)
3. **Archive Creation**: Creates compressed ZIP archives with `.cbz` extension
4. **Concurrent Processing**: Distributes work across multiple threads for optimal performance
5. **Progress Reporting**: Provides real-time feedback with colored logging

## Supported Image Formats

The tool automatically detects and includes:
- **Common formats**: JPEG, PNG, GIF, BMP, TIFF
- **Modern formats**: WebP, HEIF, AVIF
- **Future formats**: Any format with proper MIME type headers

## Output Structure

```
input/
├── Manga Title 1/
│   ├── page001.jpg
│   ├── page002.png
│   └── info.txt
└── Manga Title 2/
    ├── 01.jpg
    └── 02.jpg

output/
├── Manga Title 1.cbz
└── Manga Title 2.cbz
```

## Logging and Feedback

The tool provides professional logging with color-coded output:

- **[INFO]** - General information (blue)
- **[OK]** - Successful operations (green)  
- **[WARN]** - Warnings and skipped items (yellow)
- **[ERROR]** - Error conditions (red)

### Sample Output
```
[INFO] Starting CBZ conversion with 4 threads
[INFO] Input:  ./manga
[INFO] Output: ./cbz
[INFO] Found 199 folders to process
[WORKER 1] Processing: [Author] Title Chapter 1
[OK] [WORKER 1] Created: Title Chapter 1.cbz
[WARN] [WORKER 2] Found 2 non-image files (excluded from CBZ)
...
[INFO] Conversion completed
[INFO] Total folders:     199
[OK] Successful:        197
[WARN] Skipped:           2
[INFO] Success rate:      100.0%
```

## Performance Considerations

- **Thread Count**: Default is 4 threads. Increase for faster processing on multi-core systems
- **Memory Usage**: Each worker uses minimal memory; safe to run many threads
- **I/O Optimization**: Uses buffered operations and compression for efficiency
- **Resource Limits**: Automatically caps threads at 2× CPU cores to prevent system overload

## Error Handling

The tool handles various error conditions gracefully:

- **Missing directories**: Clear error messages with exit codes
- **Permission issues**: Skips inaccessible files with warnings
- **Corrupted files**: Uses fail-safe approach to include ambiguous files
- **Existing files**: Skips existing CBZ files to prevent overwriting
- **Individual failures**: Continues processing other folders if one fails

## Technical Details

- **Language**: Go 1.19+
- **Archive Format**: ZIP with DEFLATE compression
- **MIME Detection**: Uses Go's `http.DetectContentType()` for robust file type identification
- **Concurrency**: Worker pool pattern with bounded channels
- **Cross-Platform**: Builds for 20+ OS/architecture combinations

## Supported Platforms

The build script supports:
- **Linux**: AMD64, ARM, ARM64, PowerPC, MIPS, S390X
- **macOS**: Intel (AMD64) and Apple Silicon (ARM64)
- **FreeBSD**: AMD64, 386
- **OpenBSD**: AMD64, 386, ARM64
- **NetBSD**: AMD64, 386, ARM
- **Other**: DragonFlyBSD, Solaris, Plan 9

## Contributing

1. Fork the repository
2. Create your feature branch
3. Add tests for new functionality
4. Ensure all builds pass: `./build.sh all`
5. Submit a pull request

## License

This project is released under the MIT License.

## Troubleshooting

### Common Issues

**Q: "No image files found" error**
- Check that folders contain actual image files
- Verify file permissions are readable

**Q: High memory usage**
- Reduce thread count with `-threads` flag
- Process smaller batches of folders

**Q: CBZ files not opening in comic readers**
- Ensure input folders contain valid image files
- Some readers may need specific file ordering

**Q: Permission denied errors**
- Check read permissions on input directory
- Check write permissions on output directory
- Run with appropriate user privileges

### Performance Tips

- Use SSD storage for both input and output directories
- Set thread count to match your CPU cores (or slightly higher)
- Process folders in smaller batches for very large collections
- Close other resource-intensive applications during conversion

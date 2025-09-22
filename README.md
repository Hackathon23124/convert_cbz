# CBZ Converter

A high-performance, concurrent tool for converting folders containing images into CBZ (Comic Book Archive) files. Built in Go for speed and reliability.

![convert_cbz logo](https://jelius.dev/assets/compressed/convert_cbz.webp)

## Features

- **High Performance**: Multi-threaded processing with configurable concurrency
- **Smart Detection**: MIME-type based image detection (supports JPEG, PNG, GIF, WebP, HEIF, AVIF, and more)
- **Cross-Platform**: Runs on Linux, macOS, Windows, and various Unix systems
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
| `-dumb` | Archive all files without filtering | `false` (smart mode) |
| `-help` | Show usage information | - |
| `-version` | Show version information | - |

### Examples

Smart mode (default) - filters intelligently:
```bash
convert_cbz -input ./manga -output ./cbz
```

Dumb mode - archives everything:
```bash
convert_cbz -input ./raw_folders -output ./archives -dumb
```

High-performance processing:
```bash
convert_cbz -input /home/user/comics -output /home/user/cbz -threads 8
```

Show help:
```bash
convert_cbz -help
```

## Processing Modes

### Smart Mode (Default)
**Includes:**
- **Images**: JPEG, PNG, GIF, WebP, HEIF, AVIF, BMP, TIFF
- **Text files**: TXT, MD, NFO, INFO (metadata/descriptions)  
- **Video files**: MP4, AVI, MKV, MOV (supplementary content)
- **Any content with image/text/video MIME types**

**Excludes:**
- **System files**: .DS_Store, Thumbs.db, desktop.ini
- **Version control**: .git, .svn, .hg directories and files
- **IDE files**: .vscode, .idea, .sublime-project
- **Temporary files**: .swp, .swo, *~ backup files

### Dumb Mode (`-dumb`)
**Includes:** Everything - all files and folders are archived without any filtering whatsoever

**Use cases:**
- Preserving complete directory structures
- Archiving mixed content where filtering might remove needed files
- When you want maximum control over what gets included

## How It Works

1. **Directory Scanning**: Recursively scans each folder in the input directory
2. **Content Analysis**: 
   - **Smart Mode**: Uses MIME type analysis and filename patterns to identify useful content
   - **Dumb Mode**: Includes all files without any filtering
3. **Archive Creation**: Creates compressed ZIP archives with `.cbz` extension
4. **Concurrent Processing**: Distributes work across multiple threads for optimal performance
5. **Progress Reporting**: Provides real-time feedback with colored logging

## Supported Content Types

### Smart Mode Detection
- **Images**: Automatic MIME type detection for all formats
- **Text**: Extensions (.txt, .md, .nfo) + text/* MIME types  
- **Video**: Extensions (.mp4, .avi, .mkv) + video/* MIME types
- **Unknown**: Fail-safe inclusion for unidentifiable files

### All Formats Include
- **Image formats**: JPEG, PNG, GIF, BMP, TIFF, WebP, HEIF, AVIF
- **Video formats**: MP4, AVI, MKV, MOV, WMV, FLV, WebM
- **Text formats**: TXT, MD, NFO, INFO, README

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
[INFO] Mode: SMART - filtering files intelligently
[INFO] Found 199 folders to process
[WORKER 1] Processing: [Author] Title Chapter 1
[OK] [WORKER 1] Created: Title Chapter 1.cbz
[WARN] [WORKER 2] Found 2 files excluded by smart filtering
...
[INFO] Conversion completed
[INFO] Total folders:     199
[OK] Successful:        197
[WARN] Skipped:           2
[INFO] Files excluded:    15 (smart filtering)
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

**Q: "No files found to archive" error**
- Smart mode: Check that folders contain images, text, or video files
- Dumb mode: Verify the folder actually contains files
- Check file permissions are readable

**Q: Too many/few files being included**
- Use `-dumb` for complete archiving without filtering
- Smart mode intentionally excludes system files and VCS data
- Check the excluded files count in the final statistics

**Q: CBZ files not opening in comic readers**
- Ensure input folders contain valid image files
- Some readers may need specific file ordering
- Try both smart and dumb modes to see which works better

**Q: Permission denied errors**
- Check read permissions on input directory  
- Check write permissions on output directory
- Run with appropriate user privileges

**Q: High memory usage**
- Reduce thread count with `-threads` flag
- Process smaller batches of folders Check that folders contain actual image files
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

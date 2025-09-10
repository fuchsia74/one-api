---
applyTo: "**/*"
---
# The filesystem MCP server

These instructions describe how to efficiently work with files and directories using the filesystem MCP server. You can load this file directly into a session where the filesystem MCP server is connected.

## Detecting filesystem needs

The filesystem MCP server should be used whenever you need to interact with files and directories, including reading, writing, editing, searching, and managing file system structures. Always check allowed directories first using `mcp_mcp-filesystem_list_allowed_directories`.

## Filesystem workflows

These guidelines MUST be followed whenever working with the filesystem. There are four main workflows described below: 'Read Workflow' for examining files, 'Write Workflow' for creating content, 'Edit Workflow' for modifying existing files, and 'Search Workflow' for finding files and content.

You may re-do parts of each workflow as necessary to recover from errors. However, you must not skip any steps.

### Read workflow

The goal of the read workflow is to understand and examine file contents and directory structures.

1. **Verify access permissions**: Always start by checking which directories are accessible.
   EXAMPLE: Check allowed directories: `mcp_mcp-filesystem_list_allowed_directories({})`

2. **Explore directory structure**: Use directory listing to understand the layout before accessing specific files.
   EXAMPLE: List directory contents: `mcp_mcp-filesystem_list_directory({"path": "/project/src"})`
   EXAMPLE: Get directory tree: `mcp_mcp-filesystem_directory_tree({"path": "/project"})`

3. **Read file contents**: Choose the appropriate read method based on your needs:
   - For text files: `mcp_mcp-filesystem_read_text_file({"path": "/path/to/file.txt"})`
   - For partial content: `mcp_mcp-filesystem_read_text_file({"path": "/path/to/file.txt", "head": 50})`
   - For multiple files: `mcp_mcp-filesystem_read_multiple_files({"paths": ["/file1.txt", "/file2.txt"]})`
   - For media files: `mcp_mcp-filesystem_read_media_file({"path": "/path/to/image.jpg"})`

4. **Get file metadata**: When you need detailed file information.
   EXAMPLE: Get file info: `mcp_mcp-filesystem_get_file_info({"path": "/path/to/file.txt"})`

### Write workflow

The write workflow is for creating new files and directories.

1. **Verify target location**: Ensure the target directory exists and is accessible.
   EXAMPLE: Check directory: `mcp_mcp-filesystem_list_directory({"path": "/target/directory"})`

2. **Create directories if needed**: Use directory creation for setting up file structures.
   EXAMPLE: Create directory: `mcp_mcp-filesystem_create_directory({"path": "/new/project/structure"})`

3. **Write file content**: Create new files or completely overwrite existing ones.
   EXAMPLE: Write file: `mcp_mcp-filesystem_write_file({"path": "/path/to/newfile.txt", "content": "File content here"})`

4. **Verify creation**: Check that files were created successfully.
   EXAMPLE: Verify: `mcp_mcp-filesystem_get_file_info({"path": "/path/to/newfile.txt"})`

### Edit workflow

The edit workflow is for modifying existing files with precision.

1. **Read current content**: Always read the file first to understand its current state.
   EXAMPLE: Read file: `mcp_mcp-filesystem_read_text_file({"path": "/path/to/file.txt"})`

2. **Plan edits carefully**: Identify exact text to replace and prepare new content.

3. **Apply line-based edits**: Use the edit function for precise modifications.
   EXAMPLE: Edit file: `mcp_mcp-filesystem_edit_file({"path": "/path/to/file.txt", "edits": [{"oldText": "old content line", "newText": "new content line"}]})`

4. **Preview changes**: Use dry run to preview changes before applying.
   EXAMPLE: Preview edit: `mcp_mcp-filesystem_edit_file({"path": "/path/to/file.txt", "edits": [{"oldText": "old", "newText": "new"}], "dryRun": true})`

5. **Verify results**: Read the file again to confirm edits were applied correctly.

### Search workflow

The search workflow is for finding files, directories, and content.

1. **Search for files by name**: Use pattern-based file searching.
   EXAMPLE: Find files: `mcp_mcp-filesystem_search_files({"path": "/project", "pattern": "*.js"})`
   EXAMPLE: Search with exclusions: `mcp_mcp-filesystem_search_files({"path": "/project", "pattern": "config", "excludePatterns": ["node_modules", ".git"]})`

2. **Browse directories systematically**: Use directory listing with sorting options.
   EXAMPLE: List with sizes: `mcp_mcp-filesystem_list_directory_with_sizes({"path": "/project", "sortBy": "size"})`

3. **Navigate and organize**: Use move operations for file organization.
   EXAMPLE: Move file: `mcp_mcp-filesystem_move_file({"source": "/old/location/file.txt", "destination": "/new/location/file.txt"})`

## Tool-specific guidelines

### Reading tools
- **read_text_file**: Best for single text files. Use `head` or `tail` parameters for large files.
- **read_multiple_files**: Efficient for reading several files at once. Failed reads don't stop the operation.
- **read_media_file**: Returns base64 encoded data and MIME type for images, audio, etc.

### Writing tools
- **write_file**: Overwrites existing files completely. Use with caution.
- **edit_file**: Precise line-based edits with git-style diff output. Safer for modifications.

### Directory tools
- **list_directory**: Basic listing with file/directory distinction.
- **list_directory_with_sizes**: Includes size information, useful for space management.
- **directory_tree**: Recursive JSON structure, good for understanding project layout.
- **create_directory**: Creates nested directories in one operation.

### Search and management tools
- **search_files**: Recursive pattern matching with exclusion support.
- **move_file**: Both renaming and moving files between directories.
- **get_file_info**: Detailed metadata including permissions, timestamps, and size.

## Best practices

1. **Always check permissions first**: Use `list_allowed_directories` before attempting operations.

2. **Use appropriate read methods**: 
   - Single files: `read_text_file`
   - Multiple files: `read_multiple_files`
   - Large files: Use `head` or `tail` parameters
   - Binary files: `read_media_file`

3. **Preview before modifying**: Use `dryRun: true` with `edit_file` to see changes before applying.

4. **Handle errors gracefully**: Failed file operations return clear error messages. Check paths and permissions.

5. **Use efficient search patterns**: Include relevant exclusions to avoid searching unnecessary directories.

6. **Organize systematically**: Create directory structures before writing files.

7. **Verify operations**: Check file info or read content after write/edit operations.

## Common use cases

### Project exploration
```
// Check what directories are available
mcp_mcp-filesystem_list_allowed_directories({})

// Get project overview
mcp_mcp-filesystem_directory_tree({"path": "/project"})

// Read main files
mcp_mcp-filesystem_read_multiple_files({"paths": ["/project/README.md", "/project/package.json"]})
```

### Configuration file editing
```
// Read current config
mcp_mcp-filesystem_read_text_file({"path": "/config/app.json"})

// Preview changes
mcp_mcp-filesystem_edit_file({
  "path": "/config/app.json",
  "edits": [{"oldText": "\"debug\": false", "newText": "\"debug\": true"}],
  "dryRun": true
})

// Apply changes
mcp_mcp-filesystem_edit_file({
  "path": "/config/app.json",
  "edits": [{"oldText": "\"debug\": false", "newText": "\"debug\": true"}]
})
```

### File organization
```
// Find files to organize
mcp_mcp-filesystem_search_files({
  "path": "/downloads",
  "pattern": "*.pdf",
  "excludePatterns": ["temp", "cache"]
})

// Create organized structure
mcp_mcp-filesystem_create_directory({"path": "/organized/documents"})

// Move files
mcp_mcp-filesystem_move_file({
  "source": "/downloads/document.pdf",
  "destination": "/organized/documents/document.pdf"
})
```

### Batch file processing
```
// Find all source files
mcp_mcp-filesystem_search_files({"path": "/src", "pattern": "*.js"})

// Read multiple files for analysis
mcp_mcp-filesystem_read_multiple_files({"paths": ["/src/file1.js", "/src/file2.js"]})

// Get directory sizes for cleanup
mcp_mcp-filesystem_list_directory_with_sizes({"path": "/build", "sortBy": "size"})
```

### Safe file modification
```
// Always read first
mcp_mcp-filesystem_read_text_file({"path": "/important/config.yml"})

// Preview the change
mcp_mcp-filesystem_edit_file({
  "path": "/important/config.yml",
  "edits": [{"oldText": "version: 1.0", "newText": "version: 1.1"}],
  "dryRun": true
})

// Apply if preview looks correct
mcp_mcp-filesystem_edit_file({
  "path": "/important/config.yml",
  "edits": [{"oldText": "version: 1.0", "newText": "version: 1.1"}]
})

// Verify the change
mcp_mcp-filesystem_read_text_file({"path": "/important/config.yml", "head": 10})
```

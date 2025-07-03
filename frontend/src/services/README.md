# File Storage Service

A robust local file storage service for managing game save files with support for future pluggable adapters.

## Features

- 🔒 **Security**: File extension validation, size limits, and path traversal protection
- 📁 **Organized Storage**: Files stored in `storage/saves/{gameId}/{timestamp}.sav`
- 🔄 **Stream Support**: Returns readable streams for efficient file handling
- 🧩 **Pluggable Design**: Interface-based architecture for future adapters (e.g., S3)
- ✅ **Validation**: Comprehensive input validation and error handling
- 📦 **ZIP Support**: Validates and processes ZIP archives safely

## Quick Start

```typescript
import { saveGameFile, getLatestSave } from './fileStorage';

// Save a game file
const buffer = fs.readFileSync('game-save.sav');
const result = await saveGameFile('game-123', buffer, 'player-save.sav');
console.log('Saved to:', result.path);

// Get the latest save
const latest = await getLatestSave('game-123');
console.log('Latest save metadata:', latest.metadata);
// Stream the file data
latest.stream.pipe(response);
```

## API Reference

### Core Functions

#### `saveGameFile(gameId, buffer, originalName)`
Saves a game file to local storage.

**Parameters:**
- `gameId` (string): Unique identifier for the game
- `buffer` (Buffer): File content as a buffer
- `originalName` (string): Original filename with extension

**Returns:** `Promise<SaveFileResult>`
- `metadata`: File metadata including path, size, timestamp
- `path`: Full path to the saved file

**Example:**
```typescript
const buffer = Buffer.from('game data');
const result = await saveGameFile('my-game', buffer, 'save1.sav');
```

#### `getLatestSave(gameId)`
Retrieves the most recent save file for a game.

**Parameters:**
- `gameId` (string): Unique identifier for the game

**Returns:** `Promise<GameSaveStream>`
- `stream`: Readable stream of the file
- `metadata`: File metadata

**Example:**
```typescript
const latest = await getLatestSave('my-game');
const chunks = [];
for await (const chunk of latest.stream) {
  chunks.push(chunk);
}
const fileData = Buffer.concat(chunks);
```

### Additional Functions

#### `listGameSaves(gameId)`
Lists all save files for a game, sorted by timestamp (newest first).

#### `deleteGameSave(gameId, fileName)`
Deletes a specific save file.

## Configuration

### File Constraints
```typescript
const MAX_FILE_SIZE = 10 * 1024 * 1024; // 10MB
const ALLOWED_EXTENSIONS = ['.sav', '.zip'];
```

### Storage Structure
```
storage/
└── saves/
    └── {gameId}/
        ├── 1703123456789.sav
        ├── 1703123567890.sav
        └── 1703123678901.sav
```

## Pluggable Adapter Design

The service uses an interface-based design to support future storage adapters:

```typescript
interface FileStorageAdapter {
  saveFile(gameId: string, buffer: Buffer, originalName: string): Promise<SaveFileResult>;
  getLatestSave(gameId: string): Promise<GameSaveStream>;
  listSaves(gameId: string): Promise<SaveMetadata[]>;
  deleteSave(gameId: string, fileName: string): Promise<void>;
}
```

### Future Adapters

Examples of potential future adapters:

```typescript
// S3 Storage Adapter
export class S3FileStorageAdapter implements FileStorageAdapter {
  constructor(private s3Client: S3Client, private bucketName: string) {}
  
  async saveFile(gameId: string, buffer: Buffer, originalName: string): Promise<SaveFileResult> {
    // S3 implementation
  }
  // ... other methods
}

// Google Cloud Storage Adapter
export class GCSFileStorageAdapter implements FileStorageAdapter {
  // GCS implementation
}

// Database Storage Adapter
export class DatabaseFileStorageAdapter implements FileStorageAdapter {
  // Database blob storage implementation
}
```

### Using Different Adapters

```typescript
// Use S3 adapter
const s3Adapter = new S3FileStorageAdapter(s3Client, 'my-bucket');
const result = await s3Adapter.saveFile('game-123', buffer, 'save.sav');

// Use local adapter (default)
const localAdapter = new LocalFileStorageAdapter();
const result = await localAdapter.saveFile('game-123', buffer, 'save.sav');
```

## Security Features

### File Validation
- **Extension Whitelist**: Only `.sav` and `.zip` files allowed
- **Size Limits**: Maximum 10MB per file
- **ZIP Validation**: Checks for path traversal attacks in ZIP archives

### Path Security
- **Game ID Sanitization**: Prevents directory traversal
- **Safe File Names**: Uses timestamp-based naming to avoid conflicts

### Error Handling
```typescript
try {
  await saveGameFile('game-123', buffer, 'save.txt'); // Invalid extension
} catch (error) {
  if (error instanceof ValidationError) {
    console.log('Validation failed:', error.message);
  }
}
```

## Testing

Run the test suite:
```bash
npm test src/services/fileStorage.test.ts
```

The test suite covers:
- File saving and retrieval
- Validation edge cases
- Security scenarios
- Stream functionality
- Error handling
- Concurrent operations

## Examples

### Basic Save and Load
```typescript
import { saveGameFile, getLatestSave } from './fileStorage';

// Save a game
const gameData = JSON.stringify({ level: 5, score: 1200 });
const buffer = Buffer.from(gameData);
await saveGameFile('rpg-game', buffer, 'quicksave.sav');

// Load the latest save
const latest = await getLatestSave('rpg-game');
const chunks = [];
for await (const chunk of latest.stream) {
  chunks.push(chunk);
}
const loadedData = JSON.parse(Buffer.concat(chunks).toString());
```

### Working with ZIP Files
```typescript
import * as AdmZip from 'adm-zip';

// Create a ZIP save file
const zip = new AdmZip();
zip.addFile('save.json', Buffer.from(JSON.stringify({ progress: 100 })));
zip.addFile('screenshot.png', screenshotBuffer);

const zipBuffer = zip.toBuffer();
await saveGameFile('my-game', zipBuffer, 'complete-save.zip');
```

### Listing and Managing Saves
```typescript
// List all saves for a game
const saves = await listGameSaves('my-game');
console.log(`Found ${saves.length} saves`);

// Delete old saves (keep only latest 5)
if (saves.length > 5) {
  const oldSaves = saves.slice(5);
  for (const save of oldSaves) {
    await deleteGameSave('my-game', save.fileName);
  }
}
```

## Error Reference

### ValidationError
Thrown when input validation fails:
- Invalid file extensions
- Files exceeding size limits
- Malicious ZIP files
- Invalid game IDs

### File System Errors
- Directory creation failures
- File write/read errors
- Permission issues

## Best Practices

1. **Always validate game IDs** before using them in paths
2. **Use streams** for large files to avoid memory issues
3. **Handle errors gracefully** with proper user feedback
4. **Clean up old saves** periodically to manage disk space
5. **Consider implementing save file versioning** for complex games

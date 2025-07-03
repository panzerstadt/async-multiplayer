import { describe, it, expect, beforeEach, afterEach } from '@jest/globals';
import * as fs from 'fs';
import * as path from 'path';
import * as AdmZip from 'adm-zip';
import {
  saveGameFile,
  getLatestSave,
  listGameSaves,
  deleteGameSave,
  LocalFileStorageAdapter,
  ValidationError,
  MAX_FILE_SIZE,
  ALLOWED_EXTENSIONS,
  STORAGE_BASE_PATH,
  SaveMetadata,
  SaveFileResult,
  GameSaveStream
} from './fileStorage';

// Test configuration
const TEST_GAME_ID = 'test-game-123';
const TEST_STORAGE_PATH = path.join(STORAGE_BASE_PATH, TEST_GAME_ID);

// Helper functions
function createTestBuffer(size: number = 1024): Buffer {
  return Buffer.alloc(size, 'x');
}

function createTestZipBuffer(): Buffer {
  const zip = new (AdmZip as any)();
  zip.addFile('test.txt', Buffer.from('Hello World'));
  return zip.toBuffer();
}

function createMaliciousZipBuffer(): Buffer {
  const zip = new (AdmZip as any)();
  zip.addFile('../../../evil.txt', Buffer.from('Evil content'));
  return zip.toBuffer();
}

function cleanupTestDirectory(): void {
  if (fs.existsSync(TEST_STORAGE_PATH)) {
    fs.rmSync(TEST_STORAGE_PATH, { recursive: true, force: true });
  }
}

describe('FileStorage Service', () => {
  beforeEach(() => {
    cleanupTestDirectory();
  });

  afterEach(() => {
    cleanupTestDirectory();
  });

  describe('saveGameFile', () => {
    it('should save a .sav file successfully', async () => {
      const buffer = createTestBuffer();
      const originalName = 'test-save.sav';
      
      const result = await saveGameFile(TEST_GAME_ID, buffer, originalName);
      
      expect(result).toBeDefined();
      expect(result.metadata.gameId).toBe(TEST_GAME_ID);
      expect(result.metadata.originalName).toBe(originalName);
      expect(result.metadata.size).toBe(buffer.length);
      expect(result.metadata.extension).toBe('.sav');
      expect(result.metadata.fileName).toMatch(/^\d+\.sav$/);
      expect(fs.existsSync(result.path)).toBe(true);
    });

    it('should save a .zip file successfully', async () => {
      const buffer = createTestZipBuffer();
      const originalName = 'test-save.zip';
      
      const result = await saveGameFile(TEST_GAME_ID, buffer, originalName);
      
      expect(result).toBeDefined();
      expect(result.metadata.gameId).toBe(TEST_GAME_ID);
      expect(result.metadata.originalName).toBe(originalName);
      expect(result.metadata.extension).toBe('.zip');
      expect(fs.existsSync(result.path)).toBe(true);
    });

    it('should reject invalid file extensions', async () => {
      const buffer = createTestBuffer();
      const originalName = 'test-save.txt';
      
      await expect(saveGameFile(TEST_GAME_ID, buffer, originalName))
        .rejects
        .toThrow(ValidationError);
    });

    it('should reject files exceeding maximum size', async () => {
      const buffer = createTestBuffer(MAX_FILE_SIZE + 1);
      const originalName = 'test-save.sav';
      
      await expect(saveGameFile(TEST_GAME_ID, buffer, originalName))
        .rejects
        .toThrow(ValidationError);
    });

    it('should reject malicious zip files', async () => {
      const buffer = createMaliciousZipBuffer();
      const originalName = 'malicious.zip';
      
      await expect(saveGameFile(TEST_GAME_ID, buffer, originalName))
        .rejects
        .toThrow(ValidationError);
    });

    it('should reject empty zip files', async () => {
      const zip = new (AdmZip as any)();
      const buffer = zip.toBuffer();
      const originalName = 'empty.zip';
      
      await expect(saveGameFile(TEST_GAME_ID, buffer, originalName))
        .rejects
        .toThrow(ValidationError);
    });

    it('should reject invalid game IDs', async () => {
      const buffer = createTestBuffer();
      const originalName = 'test-save.sav';
      
      await expect(saveGameFile('', buffer, originalName))
        .rejects
        .toThrow(ValidationError);
        
      await expect(saveGameFile('../invalid', buffer, originalName))
        .rejects
        .toThrow(ValidationError);
        
      await expect(saveGameFile('game/with/slashes', buffer, originalName))
        .rejects
        .toThrow(ValidationError);
    });
  });

  describe('getLatestSave', () => {
    it('should return the latest save file', async () => {
      const buffer1 = createTestBuffer();
      const buffer2 = createTestBuffer(512);
      
      // Save first file
      await saveGameFile(TEST_GAME_ID, buffer1, 'save1.sav');
      
      // Wait a moment to ensure different timestamps
      await new Promise(resolve => setTimeout(resolve, 10));
      
      // Save second file
      await saveGameFile(TEST_GAME_ID, buffer2, 'save2.sav');
      
      const latest = await getLatestSave(TEST_GAME_ID);
      
      expect(latest).toBeDefined();
      expect(latest.metadata.gameId).toBe(TEST_GAME_ID);
      expect(latest.metadata.size).toBe(buffer2.length);
      expect(latest.stream).toBeDefined();
    });

    it('should throw error when no saves exist', async () => {
      await expect(getLatestSave(TEST_GAME_ID))
        .rejects
        .toThrow('No saves found for game');
    });

    it('should throw error for invalid game ID', async () => {
      await expect(getLatestSave(''))
        .rejects
        .toThrow(ValidationError);
    });
  });

  describe('listGameSaves', () => {
    it('should return empty array when no saves exist', async () => {
      const saves = await listGameSaves(TEST_GAME_ID);
      expect(saves).toEqual([]);
    });

    it('should return list of saves sorted by timestamp', async () => {
      const buffer1 = createTestBuffer();
      const buffer2 = createTestBuffer(512);
      
      // Save files with delay to ensure different timestamps
      await saveGameFile(TEST_GAME_ID, buffer1, 'save1.sav');
      await new Promise(resolve => setTimeout(resolve, 10));
      await saveGameFile(TEST_GAME_ID, buffer2, 'save2.sav');
      
      const saves = await listGameSaves(TEST_GAME_ID);
      
      expect(saves).toHaveLength(2);
      expect(saves[0].timestamp).toBeGreaterThan(saves[1].timestamp);
      expect(saves[0].size).toBe(buffer2.length);
      expect(saves[1].size).toBe(buffer1.length);
    });

    it('should throw error for invalid game ID', async () => {
      await expect(listGameSaves(''))
        .rejects
        .toThrow(ValidationError);
    });
  });

  describe('deleteGameSave', () => {
    it('should delete an existing save file', async () => {
      const buffer = createTestBuffer();
      const result = await saveGameFile(TEST_GAME_ID, buffer, 'test-save.sav');
      
      expect(fs.existsSync(result.path)).toBe(true);
      
      await deleteGameSave(TEST_GAME_ID, result.metadata.fileName);
      
      expect(fs.existsSync(result.path)).toBe(false);
    });

    it('should throw error when save file does not exist', async () => {
      await expect(deleteGameSave(TEST_GAME_ID, 'nonexistent.sav'))
        .rejects
        .toThrow('Save file not found');
    });

    it('should throw error for invalid game ID', async () => {
      await expect(deleteGameSave('', 'test.sav'))
        .rejects
        .toThrow(ValidationError);
    });
  });

  describe('LocalFileStorageAdapter', () => {
    it('should implement the FileStorageAdapter interface', () => {
      const adapter = new LocalFileStorageAdapter();
      
      expect(typeof adapter.saveFile).toBe('function');
      expect(typeof adapter.getLatestSave).toBe('function');
      expect(typeof adapter.listSaves).toBe('function');
      expect(typeof adapter.deleteSave).toBe('function');
    });

    it('should create directory structure when saving first file', async () => {
      const adapter = new LocalFileStorageAdapter();
      const buffer = createTestBuffer();
      
      expect(fs.existsSync(TEST_STORAGE_PATH)).toBe(false);
      
      await adapter.saveFile(TEST_GAME_ID, buffer, 'test.sav');
      
      expect(fs.existsSync(TEST_STORAGE_PATH)).toBe(true);
    });
  });

  describe('File naming and storage structure', () => {
    it('should store files in correct directory structure', async () => {
      const buffer = createTestBuffer();
      const result = await saveGameFile(TEST_GAME_ID, buffer, 'test-save.sav');
      
      expect(result.path).toMatch(new RegExp(`storage/saves/${TEST_GAME_ID}/\\d+\\.sav$`));
    });

    it('should generate unique filenames for concurrent saves', async () => {
      const buffer1 = createTestBuffer();
      const buffer2 = createTestBuffer();
      
      const [result1, result2] = await Promise.all([
        saveGameFile(TEST_GAME_ID, buffer1, 'save1.sav'),
        saveGameFile(TEST_GAME_ID, buffer2, 'save2.sav')
      ]);
      
      expect(result1.metadata.fileName).not.toBe(result2.metadata.fileName);
      expect(fs.existsSync(result1.path)).toBe(true);
      expect(fs.existsSync(result2.path)).toBe(true);
    });
  });

  describe('Stream functionality', () => {
    it('should return readable stream for saved file', async () => {
      const originalData = 'Test game save data';
      const buffer = Buffer.from(originalData);
      
      await saveGameFile(TEST_GAME_ID, buffer, 'test-save.sav');
      
      const latest = await getLatestSave(TEST_GAME_ID);
      
      // Read stream data
      const chunks: Buffer[] = [];
      for await (const chunk of latest.stream) {
        chunks.push(chunk);
      }
      
      const streamData = Buffer.concat(chunks).toString();
      expect(streamData).toBe(originalData);
    });
  });

  describe('Constants and exports', () => {
    it('should export required constants', () => {
      expect(MAX_FILE_SIZE).toBe(10 * 1024 * 1024);
      expect(ALLOWED_EXTENSIONS).toEqual(['.sav', '.zip']);
      expect(STORAGE_BASE_PATH).toBe('storage/saves');
    });

    it('should export ValidationError class', () => {
      const error = new ValidationError('Test error');
      expect(error).toBeInstanceOf(Error);
      expect(error.name).toBe('ValidationError');
      expect(error.message).toBe('Test error');
    });
  });
});

import * as fs from 'fs';
import * as path from 'path';
import { Readable } from 'stream';
import * as AdmZip from 'adm-zip';

// Configuration constants
const MAX_FILE_SIZE = 10 * 1024 * 1024; // 10MB
const ALLOWED_EXTENSIONS = ['.sav', '.zip'];
const STORAGE_BASE_PATH = 'storage/saves';

// Types and interfaces
export interface SaveMetadata {
  gameId: string;
  fileName: string;
  originalName: string;
  filePath: string;
  timestamp: number;
  size: number;
  extension: string;
  createdAt: Date;
}

export interface SaveFileResult {
  metadata: SaveMetadata;
  path: string;
}

export interface GameSaveStream {
  stream: Readable;
  metadata: SaveMetadata;
}

// Abstract interface for future pluggable adapters
export interface FileStorageAdapter {
  saveFile(gameId: string, buffer: Buffer, originalName: string): Promise<SaveFileResult>;
  getLatestSave(gameId: string): Promise<GameSaveStream>;
  listSaves(gameId: string): Promise<SaveMetadata[]>;
  deleteSave(gameId: string, fileName: string): Promise<void>;
}

// Validation utilities
class ValidationError extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'ValidationError';
  }
}

function validateFileExtension(fileName: string): void {
  const ext = path.extname(fileName).toLowerCase();
  if (!ALLOWED_EXTENSIONS.includes(ext)) {
    throw new ValidationError(
      `Invalid file extension. Allowed extensions: ${ALLOWED_EXTENSIONS.join(', ')}`
    );
  }
}

function validateFileSize(buffer: Buffer): void {
  if (buffer.length > MAX_FILE_SIZE) {
    throw new ValidationError(
      `File size exceeds maximum limit of ${MAX_FILE_SIZE / (1024 * 1024)}MB`
    );
  }
}

function validateGameId(gameId: string): void {
  if (!gameId || typeof gameId !== 'string' || gameId.trim().length === 0) {
    throw new ValidationError('Game ID is required and must be a non-empty string');
  }
  // Sanitize gameId to prevent path traversal
  if (gameId.includes('..') || gameId.includes('/') || gameId.includes('\\')) {
    throw new ValidationError('Game ID contains invalid characters');
  }
}

function validateZipFile(buffer: Buffer): void {
  try {
    const zip = new (AdmZip as any)(buffer);
    const entries = zip.getEntries();
    
    // Check if zip contains any files
    if (entries.length === 0) {
      throw new ValidationError('ZIP file is empty');
    }
    
    // Validate that zip doesn't contain suspicious files
    for (const entry of entries) {
      if (entry.entryName.includes('..') || entry.entryName.startsWith('/')) {
        throw new ValidationError('ZIP file contains potentially dangerous paths');
      }
    }
  } catch (error) {
    if (error instanceof ValidationError) {
      throw error;
    }
    throw new ValidationError('Invalid ZIP file format');
  }
}

// Local file storage implementation
export class LocalFileStorageAdapter implements FileStorageAdapter {
  private ensureDirectoryExists(dirPath: string): void {
    if (!fs.existsSync(dirPath)) {
      fs.mkdirSync(dirPath, { recursive: true });
    }
  }

  private getGameDirectoryPath(gameId: string): string {
    return path.join(STORAGE_BASE_PATH, gameId);
  }

  private generateFileName(originalName: string, timestamp: number): string {
    const ext = path.extname(originalName);
    const baseName = path.basename(originalName, ext);
    return `${timestamp}.sav`;
  }

  private createMetadata(
    gameId: string,
    fileName: string,
    originalName: string,
    filePath: string,
    buffer: Buffer,
    timestamp: number
  ): SaveMetadata {
    return {
      gameId,
      fileName,
      originalName,
      filePath,
      timestamp,
      size: buffer.length,
      extension: path.extname(originalName).toLowerCase(),
      createdAt: new Date(timestamp)
    };
  }

  async saveFile(gameId: string, buffer: Buffer, originalName: string): Promise<SaveFileResult> {
    // Validate inputs
    validateGameId(gameId);
    validateFileExtension(originalName);
    validateFileSize(buffer);
    
    // Additional validation for ZIP files
    if (path.extname(originalName).toLowerCase() === '.zip') {
      validateZipFile(buffer);
    }

    // Generate paths and ensure directory exists
    const gameDir = this.getGameDirectoryPath(gameId);
    this.ensureDirectoryExists(gameDir);
    
    const timestamp = Date.now();
    const fileName = this.generateFileName(originalName, timestamp);
    const filePath = path.join(gameDir, fileName);

    // Write file
    try {
      await fs.promises.writeFile(filePath, buffer);
    } catch (error) {
      throw new Error(`Failed to save file: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }

    // Create metadata
    const metadata = this.createMetadata(
      gameId,
      fileName,
      originalName,
      filePath,
      buffer,
      timestamp
    );

    return {
      metadata,
      path: filePath
    };
  }

  async getLatestSave(gameId: string): Promise<GameSaveStream> {
    validateGameId(gameId);
    
    const gameDir = this.getGameDirectoryPath(gameId);
    
    if (!fs.existsSync(gameDir)) {
      throw new Error(`No saves found for game: ${gameId}`);
    }

    try {
      const files = await fs.promises.readdir(gameDir);
      
      if (files.length === 0) {
        throw new Error(`No saves found for game: ${gameId}`);
      }

      // Filter and sort save files by timestamp (newest first)
      const saveFiles = files
        .filter(file => file.endsWith('.sav'))
        .map(file => {
          const timestamp = parseInt(file.replace('.sav', ''));
          return {
            fileName: file,
            timestamp: isNaN(timestamp) ? 0 : timestamp
          };
        })
        .sort((a, b) => b.timestamp - a.timestamp);

      if (saveFiles.length === 0) {
        throw new Error(`No valid save files found for game: ${gameId}`);
      }

      const latestSave = saveFiles[0];
      const filePath = path.join(gameDir, latestSave.fileName);
      
      // Get file stats for metadata
      const stats = await fs.promises.stat(filePath);
      
      // Create readable stream
      const stream = fs.createReadStream(filePath);
      
      // Create metadata (we'll need to infer the original name from the filename)
      const metadata: SaveMetadata = {
        gameId,
        fileName: latestSave.fileName,
        originalName: latestSave.fileName, // In this implementation, we use the generated name
        filePath,
        timestamp: latestSave.timestamp,
        size: stats.size,
        extension: '.sav',
        createdAt: new Date(latestSave.timestamp)
      };

      return {
        stream,
        metadata
      };
    } catch (error) {
      throw new Error(`Failed to get latest save: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  async listSaves(gameId: string): Promise<SaveMetadata[]> {
    validateGameId(gameId);
    
    const gameDir = this.getGameDirectoryPath(gameId);
    
    if (!fs.existsSync(gameDir)) {
      return [];
    }

    try {
      const files = await fs.promises.readdir(gameDir);
      const saves: SaveMetadata[] = [];

      for (const file of files) {
        if (file.endsWith('.sav')) {
          const filePath = path.join(gameDir, file);
          const stats = await fs.promises.stat(filePath);
          const timestamp = parseInt(file.replace('.sav', ''));
          
          saves.push({
            gameId,
            fileName: file,
            originalName: file,
            filePath,
            timestamp: isNaN(timestamp) ? 0 : timestamp,
            size: stats.size,
            extension: '.sav',
            createdAt: new Date(isNaN(timestamp) ? stats.mtime.getTime() : timestamp)
          });
        }
      }

      return saves.sort((a, b) => b.timestamp - a.timestamp);
    } catch (error) {
      throw new Error(`Failed to list saves: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  async deleteSave(gameId: string, fileName: string): Promise<void> {
    validateGameId(gameId);
    
    const gameDir = this.getGameDirectoryPath(gameId);
    const filePath = path.join(gameDir, fileName);
    
    if (!fs.existsSync(filePath)) {
      throw new Error(`Save file not found: ${fileName}`);
    }

    try {
      await fs.promises.unlink(filePath);
    } catch (error) {
      throw new Error(`Failed to delete save: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }
}

// Default storage adapter instance
export const fileStorage = new LocalFileStorageAdapter();

// Main API functions (backward compatibility)
export async function saveGameFile(
  gameId: string,
  buffer: Buffer,
  originalName: string
): Promise<SaveFileResult> {
  return fileStorage.saveFile(gameId, buffer, originalName);
}

export async function getLatestSave(gameId: string): Promise<GameSaveStream> {
  return fileStorage.getLatestSave(gameId);
}

// Additional utility functions
export async function listGameSaves(gameId: string): Promise<SaveMetadata[]> {
  return fileStorage.listSaves(gameId);
}

export async function deleteGameSave(gameId: string, fileName: string): Promise<void> {
  return fileStorage.deleteSave(gameId, fileName);
}

// Export types and constants for external use
export { MAX_FILE_SIZE, ALLOWED_EXTENSIONS, STORAGE_BASE_PATH, ValidationError };

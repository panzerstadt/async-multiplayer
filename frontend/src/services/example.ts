import * as fs from 'fs';
import * as path from 'path';
import * as AdmZip from 'adm-zip';
import {
  saveGameFile,
  getLatestSave,
  listGameSaves,
  deleteGameSave,
  LocalFileStorageAdapter,
  ValidationError
} from './fileStorage';

/**
 * Example: Basic save and load operations
 */
async function basicExample() {
  try {
    // Create some game data
    const gameData = {
      playerName: 'Hero',
      level: 15,
      experience: 2350,
      inventory: ['sword', 'potion', 'key'],
      position: { x: 100, y: 200 }
    };

    // Convert to buffer
    const buffer = Buffer.from(JSON.stringify(gameData, null, 2));

    // Save the game file
    console.log('Saving game...');
    const saveResult = await saveGameFile('rpg-adventure', buffer, 'quicksave.sav');
    console.log('Game saved:', saveResult.metadata);

    // Load the latest save
    console.log('Loading latest save...');
    const latest = await getLatestSave('rpg-adventure');
    console.log('Latest save metadata:', latest.metadata);

    // Read the stream data
    const chunks: Buffer[] = [];
    for await (const chunk of latest.stream) {
      chunks.push(chunk);
    }
    const loadedData = JSON.parse(Buffer.concat(chunks).toString());
    console.log('Loaded game data:', loadedData);

  } catch (error) {
    if (error instanceof ValidationError) {
      console.error('Validation error:', error.message);
    } else {
      console.error('Error:', error);
    }
  }
}

/**
 * Example: Working with ZIP archives
 */
async function zipExample() {
  try {
    // Create a ZIP file with multiple game assets
    const zip = new (AdmZip as any)();
    
    // Add game save data
    const saveData = { level: 10, score: 5000 };
    zip.addFile('save.json', Buffer.from(JSON.stringify(saveData)));
    
    // Add a screenshot (mock data)
    const screenshotData = Buffer.alloc(1024, 0xFF); // Mock image data
    zip.addFile('screenshot.png', screenshotData);
    
    // Add configuration
    const config = { difficulty: 'hard', soundEnabled: true };
    zip.addFile('config.json', Buffer.from(JSON.stringify(config)));

    // Save the ZIP file
    const zipBuffer = zip.toBuffer();
    const result = await saveGameFile('adventure-game', zipBuffer, 'complete-save.zip');
    console.log('ZIP save created:', result.metadata);

  } catch (error) {
    console.error('ZIP example error:', error);
  }
}

/**
 * Example: Managing multiple saves
 */
async function saveManagementExample() {
  try {
    const gameId = 'strategy-game';

    // Create multiple saves
    for (let i = 1; i <= 7; i++) {
      const gameState = {
        turn: i * 10,
        resources: { gold: i * 100, wood: i * 50 },
        units: i * 5
      };
      
      const buffer = Buffer.from(JSON.stringify(gameState));
      await saveGameFile(gameId, buffer, `autosave-${i}.sav`);
      
      // Small delay to ensure different timestamps
      await new Promise(resolve => setTimeout(resolve, 10));
    }

    // List all saves
    const allSaves = await listGameSaves(gameId);
    console.log(`Found ${allSaves.length} saves for ${gameId}`);

    // Keep only the latest 5 saves
    if (allSaves.length > 5) {
      const oldSaves = allSaves.slice(5); // Get saves beyond the first 5
      console.log(`Deleting ${oldSaves.length} old saves...`);
      
      for (const save of oldSaves) {
        await deleteGameSave(gameId, save.fileName);
        console.log(`Deleted: ${save.fileName}`);
      }
    }

    // List remaining saves
    const remainingSaves = await listGameSaves(gameId);
    console.log(`${remainingSaves.length} saves remaining`);

  } catch (error) {
    console.error('Save management error:', error);
  }
}

/**
 * Example: Using different storage adapters (future functionality)
 */
async function adapterExample() {
  try {
    // Use local storage (default)
    const localAdapter = new LocalFileStorageAdapter();
    const buffer = Buffer.from('Local save data');
    
    const localResult = await localAdapter.saveFile('local-game', buffer, 'local.sav');
    console.log('Local save:', localResult.metadata);

    // Future: Use S3 storage
    // const s3Adapter = new S3FileStorageAdapter(s3Client, 'game-saves-bucket');
    // const s3Result = await s3Adapter.saveFile('cloud-game', buffer, 'cloud.sav');
    // console.log('S3 save:', s3Result.metadata);

    console.log('Adapter example completed (local only for now)');

  } catch (error) {
    console.error('Adapter example error:', error);
  }
}

/**
 * Example: Error handling and validation
 */
async function errorHandlingExample() {
  console.log('Testing error handling...');

  // Test invalid file extension
  try {
    const buffer = Buffer.from('Invalid file');
    await saveGameFile('test-game', buffer, 'invalid.txt');
  } catch (error) {
    console.log('✓ Caught invalid extension error:', (error as Error).message);
  }

  // Test invalid game ID
  try {
    const buffer = Buffer.from('Valid file');
    await saveGameFile('../invalid-id', buffer, 'valid.sav');
  } catch (error) {
    console.log('✓ Caught invalid game ID error:', (error as Error).message);
  }

  // Test file too large
  try {
    const largeBuffer = Buffer.alloc(11 * 1024 * 1024); // 11MB
    await saveGameFile('test-game', largeBuffer, 'large.sav');
  } catch (error) {
    console.log('✓ Caught file too large error:', (error as Error).message);
  }

  // Test malicious ZIP
  try {
    const maliciousZip = new (AdmZip as any)();
    maliciousZip.addFile('../../../etc/passwd', Buffer.from('evil'));
    const zipBuffer = maliciousZip.toBuffer();
    await saveGameFile('test-game', zipBuffer, 'malicious.zip');
  } catch (error) {
    console.log('✓ Caught malicious ZIP error:', (error as Error).message);
  }

  console.log('Error handling tests completed');
}

/**
 * Example: Performance test with concurrent saves
 */
async function performanceExample() {
  console.log('Testing concurrent saves...');
  
  const gameId = 'perf-test';
  const saves: Promise<any>[] = [];
  
  // Create 10 concurrent saves
  for (let i = 0; i < 10; i++) {
    const data = { saveId: i, timestamp: Date.now() };
    const buffer = Buffer.from(JSON.stringify(data));
    saves.push(saveGameFile(gameId, buffer, `concurrent-${i}.sav`));
  }

  // Wait for all saves to complete
  const results = await Promise.all(saves);
  console.log(`✓ Successfully saved ${results.length} files concurrently`);

  // Verify all files exist and have unique names
  const uniqueNames = new Set(results.map(r => r.metadata.fileName));
  console.log(`✓ All ${uniqueNames.size} filenames are unique`);

  // Clean up
  const allSaves = await listGameSaves(gameId);
  for (const save of allSaves) {
    await deleteGameSave(gameId, save.fileName);
  }
  console.log('✓ Cleanup completed');
}

/**
 * Run all examples
 */
export async function runAllExamples() {
  console.log('=== File Storage Service Examples ===\n');

  console.log('1. Basic Example:');
  await basicExample();
  console.log('');

  console.log('2. ZIP Example:');
  await zipExample();
  console.log('');

  console.log('3. Save Management Example:');
  await saveManagementExample();
  console.log('');

  console.log('4. Adapter Example:');
  await adapterExample();
  console.log('');

  console.log('5. Error Handling Example:');
  await errorHandlingExample();
  console.log('');

  console.log('6. Performance Example:');
  await performanceExample();
  console.log('');

  console.log('All examples completed!');
}

// Export individual examples for selective testing
export {
  basicExample,
  zipExample,
  saveManagementExample,
  adapterExample,
  errorHandlingExample,
  performanceExample
};

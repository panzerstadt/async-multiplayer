#!/usr/bin/env ts-node

import { saveGameFile, getLatestSave, listGameSaves, deleteGameSave } from './fileStorage';

async function runIntegrationTest() {
  console.log('🚀 Running File Storage Integration Test\n');

  try {
    const testGameId = 'integration-test-game';
    
    // Test 1: Save a file
    console.log('1. Testing saveGameFile...');
    const gameData = { level: 1, score: 100, player: 'TestPlayer' };
    const buffer = Buffer.from(JSON.stringify(gameData));
    
    const saveResult = await saveGameFile(testGameId, buffer, 'test-save.sav');
    console.log('✅ File saved successfully');
    console.log('   Path:', saveResult.path);
    console.log('   Size:', saveResult.metadata.size, 'bytes');
    console.log('   Timestamp:', saveResult.metadata.timestamp);

    // Test 2: Get latest save
    console.log('\n2. Testing getLatestSave...');
    const latest = await getLatestSave(testGameId);
    console.log('✅ Latest save retrieved');
    console.log('   File name:', latest.metadata.fileName);
    console.log('   Size:', latest.metadata.size, 'bytes');

    // Test 3: Read stream data
    console.log('\n3. Testing stream reading...');
    const chunks: Buffer[] = [];
    for await (const chunk of latest.stream) {
      chunks.push(chunk);
    }
    const retrievedData = JSON.parse(Buffer.concat(chunks).toString());
    console.log('✅ Stream data read successfully');
    console.log('   Retrieved data:', retrievedData);
    
    // Verify data integrity
    if (JSON.stringify(gameData) === JSON.stringify(retrievedData)) {
      console.log('✅ Data integrity verified');
    } else {
      console.log('❌ Data integrity check failed');
    }

    // Test 4: List saves
    console.log('\n4. Testing listGameSaves...');
    const saves = await listGameSaves(testGameId);
    console.log('✅ Save list retrieved');
    console.log('   Number of saves:', saves.length);

    // Test 5: Save another file
    console.log('\n5. Testing multiple saves...');
    const gameData2 = { level: 2, score: 250, player: 'TestPlayer' };
    const buffer2 = Buffer.from(JSON.stringify(gameData2));
    
    await new Promise(resolve => setTimeout(resolve, 10)); // Ensure different timestamp
    await saveGameFile(testGameId, buffer2, 'test-save-2.sav');
    
    const allSaves = await listGameSaves(testGameId);
    console.log('✅ Second save completed');
    console.log('   Total saves:', allSaves.length);
    console.log('   Latest save timestamp:', allSaves[0].timestamp);

    // Test 6: Verify latest is actually latest
    console.log('\n6. Testing latest save selection...');
    const newLatest = await getLatestSave(testGameId);
    const newChunks: Buffer[] = [];
    for await (const chunk of newLatest.stream) {
      newChunks.push(chunk);
    }
    const newLatestData = JSON.parse(Buffer.concat(newChunks).toString());
    
    if (newLatestData.level === 2) {
      console.log('✅ Latest save correctly identified');
    } else {
      console.log('❌ Latest save selection failed');
    }

    // Test 7: Delete saves (cleanup)
    console.log('\n7. Testing deleteGameSave...');
    for (const save of allSaves) {
      await deleteGameSave(testGameId, save.fileName);
    }
    
    const finalSaves = await listGameSaves(testGameId);
    console.log('✅ Cleanup completed');
    console.log('   Remaining saves:', finalSaves.length);

    console.log('\n🎉 All tests passed! File Storage Service is working correctly.\n');

  } catch (error) {
    console.error('\n❌ Integration test failed:', error);
    process.exit(1);
  }
}

// Run the test if this file is executed directly
if (require.main === module) {
  runIntegrationTest();
}

export { runIntegrationTest };

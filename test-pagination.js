// Simple test script to verify pagination functionality
// This script will test the pagination by checking if the correct number of rows are displayed

const puppeteer = require('puppeteer');

async function testPagination() {
  const browser = await puppeteer.launch({ headless: false });
  const page = await browser.newPage();
  
  try {
    // Navigate to the application
    await page.goto('http://localhost:3001');
    
    // Wait for the page to load
    await page.waitForTimeout(2000);
    
    // Check if we need to login
    const loginButton = await page.$('button[type="submit"]');
    if (loginButton) {
      console.log('Login required, logging in...');
      await page.type('input[name="username"]', 'root');
      await page.type('input[name="password"]', '123456');
      await page.click('button[type="submit"]');
      await page.waitForTimeout(2000);
    }
    
    // Navigate to channels page
    await page.goto('http://localhost:3001/channels');
    await page.waitForTimeout(3000);
    
    // Wait for the table to load
    await page.waitForSelector('table tbody tr', { timeout: 10000 });
    
    // Count initial rows
    const initialRows = await page.$$eval('table tbody tr', rows => rows.length);
    console.log(`Initial rows displayed: ${initialRows}`);
    
    // Find and click the page size selector
    const pageSizeSelector = await page.$('[role="combobox"]');
    if (pageSizeSelector) {
      await pageSizeSelector.click();
      await page.waitForTimeout(1000);
      
      // Select 10 rows per page
      const option10 = await page.$('[role="option"][data-value="10"]');
      if (option10) {
        await option10.click();
        await page.waitForTimeout(2000);
        
        // Count rows after changing page size
        const rowsAfter10 = await page.$$eval('table tbody tr', rows => rows.length);
        console.log(`Rows displayed after selecting 10 per page: ${rowsAfter10}`);
        
        if (rowsAfter10 <= 10) {
          console.log('✅ Page size change to 10 works correctly');
        } else {
          console.log('❌ Page size change to 10 failed - showing more than 10 rows');
        }
      }
      
      // Try changing to 20 rows per page
      await pageSizeSelector.click();
      await page.waitForTimeout(1000);
      
      const option20 = await page.$('[role="option"][data-value="20"]');
      if (option20) {
        await option20.click();
        await page.waitForTimeout(2000);
        
        const rowsAfter20 = await page.$$eval('table tbody tr', rows => rows.length);
        console.log(`Rows displayed after selecting 20 per page: ${rowsAfter20}`);
        
        if (rowsAfter20 <= 20) {
          console.log('✅ Page size change to 20 works correctly');
        } else {
          console.log('❌ Page size change to 20 failed - showing more than 20 rows');
        }
      }
    }
    
    console.log('Pagination test completed');
    
  } catch (error) {
    console.error('Test failed:', error);
  } finally {
    await browser.close();
  }
}

// Run the test
testPagination();

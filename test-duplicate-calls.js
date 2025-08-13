// Test script to verify that duplicate API calls are fixed
// This will monitor network requests and count API calls

const puppeteer = require('puppeteer');

async function testDuplicateCalls() {
  const browser = await puppeteer.launch({ headless: false });
  const page = await browser.newPage();
  
  // Track API calls
  const apiCalls = [];
  
  page.on('request', request => {
    const url = request.url();
    if (url.includes('/api/channel/') || url.includes('/api/user/') || url.includes('/api/token/')) {
      apiCalls.push({
        url: url,
        timestamp: Date.now()
      });
      console.log(`API Call: ${url}`);
    }
  });
  
  try {
    // Navigate to the application
    await page.goto('http://localhost:3001');
    await page.waitForTimeout(2000);
    
    // Login if needed
    const loginButton = await page.$('button[type="submit"]');
    if (loginButton) {
      console.log('Logging in...');
      await page.type('input[name="username"]', 'root');
      await page.type('input[name="password"]', '123456');
      await page.click('button[type="submit"]');
      await page.waitForTimeout(2000);
    }
    
    // Test Channels page
    console.log('\n=== Testing Channels Page ===');
    await page.goto('http://localhost:3001/channels');
    await page.waitForTimeout(3000);
    
    // Clear previous API calls
    apiCalls.length = 0;
    
    // Change page size
    console.log('Changing page size to 20...');
    const pageSizeSelector = await page.$('[role="combobox"]');
    if (pageSizeSelector) {
      await pageSizeSelector.click();
      await page.waitForTimeout(1000);
      
      const option20 = await page.$('[role="option"][data-value="20"]');
      if (option20) {
        await option20.click();
        await page.waitForTimeout(3000);
        
        // Count API calls in the last 3 seconds
        const recentCalls = apiCalls.filter(call => 
          call.url.includes('/api/channel/') && 
          Date.now() - call.timestamp < 3000
        );
        
        console.log(`API calls made: ${recentCalls.length}`);
        recentCalls.forEach(call => console.log(`  - ${call.url}`));
        
        if (recentCalls.length === 1) {
          console.log('✅ Channels page: No duplicate API calls');
        } else {
          console.log(`❌ Channels page: ${recentCalls.length} API calls (should be 1)`);
        }
      }
    }
    
    // Test Users page
    console.log('\n=== Testing Users Page ===');
    await page.goto('http://localhost:3001/users');
    await page.waitForTimeout(3000);
    
    // Clear previous API calls
    apiCalls.length = 0;
    
    // Change page size
    console.log('Changing page size to 20...');
    const userPageSizeSelector = await page.$('[role="combobox"]');
    if (userPageSizeSelector) {
      await userPageSizeSelector.click();
      await page.waitForTimeout(1000);
      
      const option20 = await page.$('[role="option"][data-value="20"]');
      if (option20) {
        await option20.click();
        await page.waitForTimeout(3000);
        
        // Count API calls in the last 3 seconds
        const recentCalls = apiCalls.filter(call => 
          call.url.includes('/api/user/') && 
          Date.now() - call.timestamp < 3000
        );
        
        console.log(`API calls made: ${recentCalls.length}`);
        recentCalls.forEach(call => console.log(`  - ${call.url}`));
        
        if (recentCalls.length === 1) {
          console.log('✅ Users page: No duplicate API calls');
        } else {
          console.log(`❌ Users page: ${recentCalls.length} API calls (should be 1)`);
        }
      }
    }
    
    console.log('\nTest completed!');
    
  } catch (error) {
    console.error('Test failed:', error);
  } finally {
    await browser.close();
  }
}

// Run the test
testDuplicateCalls();

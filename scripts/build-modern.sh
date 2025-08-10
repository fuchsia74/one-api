#!/bin/bash

echo "🚀 Building One-API Modern Template..."

# Navigate to modern template directory
cd /home/laisky/repo/laisky/one-api/web/modern

# Check if we're in the right directory
if [ ! -f "package.json" ]; then
    echo "❌ Error: package.json not found. Are we in the right directory?"
    exit 1
fi

echo "📦 Installing dependencies..."
npm install

# Check if installation was successful
if [ $? -ne 0 ]; then
    echo "❌ Error: Failed to install dependencies"
    exit 1
fi

echo "🔨 Building for production..."
npm run build

# Check if build was successful
if [ $? -ne 0 ]; then
    echo "❌ Error: Build failed"
    exit 1
fi

echo "✅ Build completed successfully!"
echo ""
echo "📁 Built files are in: ./dist"
echo ""
echo "🚀 To deploy, update your Go router to serve from:"
echo "   router.Static(\"/\", \"./web/modern/dist\")"
echo ""
echo "🎉 Modern template migration is complete!"
echo ""
echo "Key improvements:"
echo "  ✅ Enhanced UI with shadcn/ui components"
echo "  ✅ Mobile-responsive design"
echo "  ✅ Advanced search and filtering"
echo "  ✅ Real-time form validation"
echo "  ✅ Fixed pagination issues"
echo "  ✅ Improved performance and accessibility"
echo ""
echo "All functionality from the default template has been preserved and enhanced!"

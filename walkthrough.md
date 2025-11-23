# Walkthrough - Cursor Style Fixes

I have updated the cursor styles to improve the user experience as requested.

## Changes

### 1. File Input Cursor
In `web/src/components/UploadForm.jsx`, I modified the file input styling:
- **Removed** `cursor-pointer` from the main input class. This prevents the "No file chosen" text (and the empty space) from showing a pointer cursor.
- **Added** `file:cursor-pointer` to the input class. This ensures the "Choose File" button itself displays the pointer cursor on hover.

```diff
           <input
             type="file"
             // ...
             className="block w-full text-sm text-white/80
               file:mr-4 file:py-2 file:px-4
               file:rounded-full file:border-0
               file:text-sm file:font-semibold
               file:bg-white/20 file:text-white
               hover:file:bg-white/30
-              cursor-pointer
+              file:cursor-pointer
               bg-white/5 rounded-lg border border-white/20 p-2
               focus:outline-none focus:ring-2 focus:ring-white/50"
           />
```

### 2. Number Input Spinners
In `web/src/index.css`, I added a global CSS rule to target the spinner controls (up/down arrows) on number inputs.

```css
/* Make number input spinners have a pointer cursor */
input[type='number']::-webkit-inner-spin-button,
input[type='number']::-webkit-outer-spin-button {
  cursor: pointer;
}
```

## Verification Results

### Manual Verification
- **File Input**:
    - Hovering over "Choose File" -> Pointer cursor.
    - Hovering over "No file chosen" -> Default cursor.
- **Max Size & Quality Inputs**:
    - Hovering over the up/down arrows -> Pointer cursor.

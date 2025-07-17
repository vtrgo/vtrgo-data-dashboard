// console/src/components/ConfigDrawer.tsx
import { useState } from "react"
import {
  Drawer,
  DrawerContent,
  DrawerHeader,
  DrawerTitle,
  DrawerDescription,
  DrawerClose,
} from "@/components/ui/drawer"
import { Button } from "@/components/ui/button"
import { useTheme } from "@/components/ui/theme-provider"

type Props = {
  open: boolean
  onOpenChange: (open: boolean) => void
  themes: string[]
  themeIndex: number
  setThemeIndex: (index: number) => void
  enableTheming: boolean
  setEnableTheming: (enabled: boolean) => void
}

export function ConfigDrawer({
  open,
  onOpenChange,
  themes,
  themeIndex,
  setThemeIndex,
  enableTheming,
  setEnableTheming,
}: Props) {
  const [selectedFile, setSelectedFile] = useState<File | null>(null)
  const [uploadStatus, setUploadStatus] = useState<
    "idle" | "uploading" | "success" | "error"
  >("idle")
  const [uploadMessage, setUploadMessage] = useState("")
  const { theme, setTheme } = useTheme()

  const handleFileChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    if (event.target.files && event.target.files[0]) {
      const file = event.target.files[0]
      if (file.type === "text/csv") {
        setSelectedFile(file)
        setUploadStatus("idle")
        setUploadMessage("")
      } else {
        setSelectedFile(null)
        setUploadStatus("error")
        setUploadMessage("Please select a valid CSV file.")
      }
    }
  }

  const handleUpload = async () => {
    if (!selectedFile) return

    setUploadStatus("uploading")
    setUploadMessage("")

    const formData = new FormData()
    formData.append("file", selectedFile)

    try {
      // This endpoint needs to be created on your backend.
      // It should be configured to save the file to the './shared' directory.
      const response = await fetch("/api/upload-csv", {
        method: "POST",
        body: formData,
      })

      if (!response.ok) {
        const errorData = await response.json()
        throw new Error(errorData.message || "File upload failed")
      }

      const result = await response.json()
      setUploadStatus("success")
      setUploadMessage(result.message || "File uploaded successfully!")
      setSelectedFile(null) // Clear file input after successful upload
    } catch (error) {
      setUploadStatus("error")
      setUploadMessage(error instanceof Error ? error.message : "An unknown error occurred.")
    }
  }

  return (
    <Drawer open={open} onOpenChange={onOpenChange}>
      <DrawerContent className="p-6 max-w-md ml-auto bg-background border-l shadow-lg">
        <DrawerHeader>
          <DrawerTitle className="text-foreground">Configuration</DrawerTitle>
          <DrawerDescription className="text-muted-foreground">
            View and modify system settings.
          </DrawerDescription>
        </DrawerHeader>

        <div className="space-y-4 py-4">
          {/* Theme Selector */}
          <div className="flex items-center justify-between">
            <label className="text-sm font-medium text-foreground">Theme</label>
            <select
              className="border rounded-md px-2 py-1 bg-background text-sm text-foreground"
              value={theme}
              onChange={e => setTheme(e.target.value as "light" | "dark" | "system")}
            >
              <option value="light">Light</option>
              <option value="dark">Dark</option>
            </select>
          </div>

          <div className="flex items-center justify-between">
            <label className="text-sm font-medium text-foreground">Enable Title Theming</label>
            <input
              type="checkbox"
              checked={enableTheming}
              onChange={(e) => setEnableTheming(e.target.checked)}
              className="w-4 h-4 accent-primary"
            />
          </div>
          <div className="text-sm italic text-muted-foreground">
            {themes.length} title themes loaded
          </div>
          {enableTheming && (
            <select
              className="w-full border rounded-md px-3 py-2 bg-background text-sm text-foreground shadow-sm"
              value={themeIndex}
              onChange={(e) => setThemeIndex(Number(e.target.value))}
            >
              {themes.map((theme, index) => (
                <option key={theme} value={index}>
                  {theme}
                </option>
              ))}
            </select>
          )}

          {/* File Upload Section */}
          <div className="space-y-2 pt-4 border-t border-border">
            <label className="text-sm font-medium text-foreground">Upload Project CSV</label>
            <div className="text-sm italic text-muted-foreground">
              Upload a CSV file to the project's shared folder.
            </div>
            <div className="flex items-center space-x-2">
              <Button asChild variant="outline">
                <label htmlFor="csvUploader" className="cursor-pointer text-foreground">
                  Choose File
                </label>
              </Button>
              <input
                type="file"
                id="csvUploader"
                accept=".csv"
                onChange={handleFileChange}
                className="sr-only"
              />
              <span className="text-sm text-muted-foreground truncate">
                {selectedFile ? selectedFile.name : "No file selected"}
              </span>
            </div>
            <Button
              onClick={handleUpload}
              disabled={!selectedFile || uploadStatus === "uploading"}
              className="w-full"
            >
              {uploadStatus === "uploading" ? "Uploading..." : "Upload"}
            </Button>
            {uploadMessage && (
              <div className={`text-sm ${uploadStatus === 'error' ? 'text-destructive' : 'text-success'}`}>{uploadMessage}</div>
            )}
          </div>

          <div className="text-sm text-muted-foreground">
            <strong className="text-foreground">PLC Source:</strong> ethernet-ip
          </div>
          <div className="text-sm text-muted-foreground">
            <strong className="text-foreground">Poll Interval:</strong> 1000 ms
          </div>
          <div className="text-sm text-muted-foreground">
            <strong className="text-foreground">Influx Bucket:</strong> status_data
          </div>
        </div>

        <DrawerClose asChild>
          <Button variant="outline" className="mt-4">Close</Button>
        </DrawerClose>
      </DrawerContent>
    </Drawer>
  )
}

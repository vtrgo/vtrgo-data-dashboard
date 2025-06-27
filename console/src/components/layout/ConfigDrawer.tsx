// console/src/components/ConfigDrawer.tsx
import {
  Drawer,
  DrawerContent,
  DrawerHeader,
  DrawerTitle,
  DrawerDescription,
  DrawerClose,
} from "@/components/ui/drawer"
import { Button } from "@/components/ui/button"

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
  return (
    <Drawer open={open} onOpenChange={onOpenChange}>
      <DrawerContent className="p-6 max-w-md ml-auto bg-white border-l shadow-lg">
        <DrawerHeader>
          <DrawerTitle>Configuration</DrawerTitle>
          <DrawerDescription>
            View and modify system settings.
          </DrawerDescription>
        </DrawerHeader>

        {/* Replace with real config */}
        <div className="space-y-4 py-4">
          <div className="flex items-center justify-between">
            <label className="text-sm font-medium">Enable Theming</label>
            <input
              type="checkbox"
              checked={enableTheming}
              onChange={(e) => setEnableTheming(e.target.checked)}
              className="w-4 h-4 accent-black"
            />
          </div>
          <div className="text-sm italic text-muted-foreground">
            {themes.length} themes loaded
          </div>
          {enableTheming && (
            <select
              className="w-full border rounded-md px-3 py-2 bg-background text-sm shadow-sm"
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
          <div className="text-sm">
            <strong>PLC Source:</strong> ethernet-ip
          </div>
          <div className="text-sm">
            <strong>Poll Interval:</strong> 1000 ms
          </div>
          <div className="text-sm">
            <strong>Influx Bucket:</strong> status_data
          </div>
        </div>

        <DrawerClose asChild>
          <Button variant="outline" className="mt-4">Close</Button>
        </DrawerClose>
      </DrawerContent>
    </Drawer>
  )
}

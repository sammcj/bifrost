'use client'

import CoreSettingsList from '@/components/config/core-settings-list'

export default function ConfigPage() {
  return (
    <div className="bg-background space-y-6">
      {/* Page Header */}
      <div>
        <h1 className="text-3xl font-bold">Configuration</h1>
        <p className="text-muted-foreground mt-2">Configure AI providers, API keys, and system settings for your Bifrost instance.</p>
      </div>

      <CoreSettingsList />
    </div>
  )
}

"use client"

import { IS_ENTERPRISE } from "@/lib/constants/config"
import { redirect } from "next/navigation"
import ProxyView from "../views/proxyView"

export default function ProxyPage() {
  if (!IS_ENTERPRISE) {
    redirect("/workspace/config/client-settings")
  }

  return (
    <div className="mx-auto flex w-full max-w-7xl">
      <ProxyView />
    </div>
  )
}


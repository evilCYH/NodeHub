"use client"

import * as React from "react"
import { type Icon } from "@tabler/icons-react"

import {
  SidebarGroup,
  SidebarGroupContent,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/src/components/ui/sidebar"
import { useRouter } from "@/src/router"

export function NavSecondary({
  items,
  onItemClick: _onItemClick,
  ...props
}: {
  items: {
    title: string
    url: string
    icon: Icon
    onClick?: () => void
  }[]
  onItemClick?: (item: { title: string; url: string; icon: Icon; onClick?: () => void }) => void
} & React.ComponentPropsWithoutRef<typeof SidebarGroup>) {
  const { currentPath, navigate } = useRouter()

  return (
    <SidebarGroup {...props}>
      <SidebarGroupContent>
        <SidebarMenu>
          {items.map((item) => {
            const isExternal = item.url.startsWith('http')
            const isActive = !isExternal && currentPath === item.url

            if (item.onClick) {
              return (
                <SidebarMenuItem key={item.title}>
                  <SidebarMenuButton onClick={item.onClick} isActive={isActive}>
                    <item.icon />
                    <span>{item.title}</span>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              )
            }

            if (isExternal) {
              return (
                <SidebarMenuItem key={item.title}>
                  <SidebarMenuButton asChild isActive={isActive}>
                    <a href={item.url} target="_blank" rel="noopener noreferrer">
                      <item.icon />
                      <span>{item.title}</span>
                    </a>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              )
            }

            return (
              <SidebarMenuItem key={item.title}>
                <SidebarMenuButton
                  isActive={isActive}
                  onClick={() => navigate(item.url)}
                >
                  <item.icon />
                  <span>{item.title}</span>
                </SidebarMenuButton>
              </SidebarMenuItem>
            )
          })}
        </SidebarMenu>
      </SidebarGroupContent>
    </SidebarGroup>
  )
}

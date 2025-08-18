"use client"

import { useState, useEffect } from "react"
import { usePathname } from "next/navigation"
import { type LucideIcon, ChevronRight } from "lucide-react"

import {
  SidebarGroup,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarMenuSub,
  SidebarMenuSubButton,
  SidebarMenuSubItem,
} from "@/components/ui/sidebar"
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible"

interface NavItem {
  title: string
  url: string
  icon?: LucideIcon
  isActive?: boolean
  items?: {
    title: string
    url: string
    isActive?: boolean
  }[]
}

export function NavMain({
  items,
}: {
  items: NavItem[]
}) {
  const pathname = usePathname()
  const [openItems, setOpenItems] = useState<Set<string>>(new Set())

  // Function to check if any subitem is active
  const isParentActive = (item: NavItem) => {
    if (!item.items) return false
    return item.items.some(subItem => pathname === subItem.url)
  }

  // Function to check if item should be open
  const shouldBeOpen = (item: NavItem) => {
    return isParentActive(item) || openItems.has(item.title)
  }

  // Auto-expand parent items when their subitems are active
  useEffect(() => {
    const newOpenItems = new Set(openItems)

    items.forEach(item => {
      if (item.items && isParentActive(item)) {
        newOpenItems.add(item.title)
      }
    })

    if (newOpenItems.size !== openItems.size ||
        [...newOpenItems].some(item => !openItems.has(item))) {
      setOpenItems(newOpenItems)
    }
  }, [pathname, items])

  const toggleItem = (title: string) => {
    setOpenItems(prev => {
      const newSet = new Set(prev)
      if (newSet.has(title)) {
        newSet.delete(title)
      } else {
        newSet.add(title)
      }
      return newSet
    })
  }

  return (
    <SidebarGroup>
      <SidebarGroupLabel>Sistema</SidebarGroupLabel>
      <SidebarMenu>
        {items.map((item) => (
          <SidebarMenuItem key={item.title}>
            {item.items ? (
              <Collapsible
                open={shouldBeOpen(item)}
                onOpenChange={() => toggleItem(item.title)}
              >
                <CollapsibleTrigger asChild>
                  <SidebarMenuButton isActive={item.isActive || isParentActive(item)}>
                    {item.icon && <item.icon />}
                    <span>{item.title}</span>
                    <ChevronRight className={`ml-auto h-4 w-4 transition-transform ${
                      shouldBeOpen(item) ? 'rotate-90' : ''
                    }`} />
                  </SidebarMenuButton>
                </CollapsibleTrigger>
                <CollapsibleContent>
                  <SidebarMenuSub>
                    {item.items.map((subItem) => (
                      <SidebarMenuSubItem key={subItem.title}>
                        <SidebarMenuSubButton asChild isActive={pathname === subItem.url}>
                          <a href={subItem.url}>
                            <span>{subItem.title}</span>
                          </a>
                        </SidebarMenuSubButton>
                      </SidebarMenuSubItem>
                    ))}
                  </SidebarMenuSub>
                </CollapsibleContent>
              </Collapsible>
            ) : (
              <SidebarMenuButton asChild isActive={pathname === item.url}>
                <a href={item.url}>
                  {item.icon && <item.icon />}
                  <span>{item.title}</span>
                </a>
              </SidebarMenuButton>
            )}
          </SidebarMenuItem>
        ))}
      </SidebarMenu>
    </SidebarGroup>
  )
}

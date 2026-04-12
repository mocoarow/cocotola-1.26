import { BookOpenIcon, GlobeIcon, LogOutIcon } from "lucide-react";
import { Form, Link, Outlet, redirect, useLoaderData, useLocation } from "react-router";
import { Button } from "~/components/ui/button";
import { Separator } from "~/components/ui/separator";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarInset,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarProvider,
  SidebarSeparator,
  SidebarTrigger,
} from "~/components/ui/sidebar";
import { fetchWithIdToken } from "~/lib/api/fetch.server";
import { requireAuth } from "~/lib/auth/require-auth.server";
import { destroySession, getSession } from "~/lib/auth/session.server";
import type { Route } from "./+types/workbooks";

export async function loader({ request }: Route.LoaderArgs) {
  console.info("[workbooks] loader called");
  const { accessToken } = await requireAuth(request);

  const authUrl = process.env.AUTH_BASE_URL;
  if (!authUrl) {
    throw new Error("AUTH_BASE_URL environment variable is required");
  }

  const meUrl = `${authUrl}/api/v1/auth/me`;
  console.info(`[workbooks] fetching user info: url=${meUrl}`);

  const response = await fetchWithIdToken("cocotola-auth", meUrl, {
    headers: { Authorization: `Bearer ${accessToken}` },
  });

  if (response.status === 401) {
    console.info("[workbooks] /auth/me returned 401, destroying session");
    const session = await getSession(request);
    throw redirect("/login", { headers: { "Set-Cookie": await destroySession(session) } });
  }

  if (!response.ok) {
    console.error(`[workbooks] /auth/me failed: status=${response.status}`);
    return { user: null };
  }

  const user = (await response.json()) as {
    userId: string;
    loginId: string;
    organizationName: string;
  };
  console.info(`[workbooks] user loaded: userId=${user.userId}, loginId=${user.loginId}`);
  return { user };
}

const navItems = [
  {
    title: "My Workbooks",
    href: "/workbooks",
    icon: BookOpenIcon,
    disabled: false,
  },
  {
    title: "Public",
    href: "/workbooks/public",
    icon: GlobeIcon,
    disabled: true,
  },
];

export default function WorkbooksLayout() {
  const { user } = useLoaderData<typeof loader>();
  const location = useLocation();

  return (
    <SidebarProvider>
      <Sidebar>
        <SidebarHeader>
          <Link to="/workbooks" className="flex items-center gap-2 px-2 py-1">
            <BookOpenIcon className="size-5" />
            <span className="text-lg font-bold">Cocotola</span>
          </Link>
        </SidebarHeader>
        <SidebarSeparator />
        <SidebarContent>
          <SidebarGroup>
            <SidebarGroupLabel>Workbooks</SidebarGroupLabel>
            <SidebarGroupContent>
              <SidebarMenu>
                {navItems.map((item) => (
                  <SidebarMenuItem key={item.title}>
                    <SidebarMenuButton
                      isActive={!item.disabled && location.pathname === item.href}
                      render={
                        item.disabled ? (
                          <span className="opacity-50 cursor-not-allowed" />
                        ) : (
                          <Link to={item.href} />
                        )
                      }
                      tooltip={item.disabled ? `${item.title} (coming soon)` : item.title}
                    >
                      <item.icon className="size-4" />
                      <span>{item.title}</span>
                      {item.disabled && (
                        <span className="ml-auto rounded-full bg-muted px-2 py-0.5 text-[10px] text-muted-foreground">
                          soon
                        </span>
                      )}
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                ))}
              </SidebarMenu>
            </SidebarGroupContent>
          </SidebarGroup>
        </SidebarContent>
        <SidebarSeparator />
        <SidebarFooter>
          {user && (
            <div className="flex items-center justify-between gap-2 px-2">
              <div className="min-w-0">
                <p className="truncate text-sm font-medium">{user.loginId}</p>
                <p className="truncate text-xs text-muted-foreground">{user.organizationName}</p>
              </div>
              <Form method="post" action="/logout">
                <Button variant="ghost" size="icon-sm" type="submit">
                  <LogOutIcon className="size-4" />
                  <span className="sr-only">Logout</span>
                </Button>
              </Form>
            </div>
          )}
        </SidebarFooter>
      </Sidebar>
      <SidebarInset>
        <header className="flex h-12 items-center gap-2 border-b px-4">
          <SidebarTrigger />
          <Separator orientation="vertical" className="mx-1 h-4" />
        </header>
        <div className="flex-1 overflow-auto p-6">
          <Outlet />
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}

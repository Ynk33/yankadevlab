import { type FormEvent, useEffect, useState } from "react";
import { Navigate, useSearchParams } from "react-router";
import { useAuth } from "@/hooks/use-auth";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

/**
 * Returns `rd` only if it targets an https sibling subdomain of the current host (open-redirect guard).
 */
function getRedirectTarget(rd: string | null): string | undefined {
  const { hostname } = window.location;
  const dot = hostname.indexOf(".");
  if (!rd || dot === -1) return undefined;

  try {
    const url = new URL(rd);
    if (
      url.protocol === "https:" &&
      url.hostname.endsWith(hostname.slice(dot))
    ) {
      return url.href;
    }
  } catch {
    return undefined;
  }
  return undefined;
}

export default function LoginPage() {
  const { isAuthenticated, login, logout, refreshAccessToken } = useAuth();
  const [searchParams] = useSearchParams();
  const redirectTarget = getRedirectTarget(searchParams.get("rd"));
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (!isAuthenticated || !redirectTarget) return;
    refreshAccessToken().then((token) => {
      if (token) window.location.assign(redirectTarget);
      else logout();
    });
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  if (isAuthenticated) {
    return redirectTarget ? null : <Navigate to="/" replace />;
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setError(null);
    setLoading(true);

    try {
      await login(email, password);
      if (redirectTarget) window.location.assign(redirectTarget);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Login failed");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-background px-4">
      <Card className="w-full max-w-sm">
        <CardHeader className="text-center">
          <CardTitle className="text-2xl">YankaDevLab</CardTitle>
          <CardDescription>Sign in to your dashboard</CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="grid gap-4">
            <div className="grid gap-2">
              <Label htmlFor="email">Email</Label>
              <Input
                id="email"
                type="email"
                required
                autoComplete="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="password">Password</Label>
              <Input
                id="password"
                type="password"
                required
                autoComplete="current-password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
              />
            </div>
            {error && <p className="text-sm text-destructive">{error}</p>}
            <Button type="submit" className="w-full" disabled={loading}>
              {loading ? "Signing in..." : "Sign in"}
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}

import { Button } from "@/components/ui/button";
import { Link } from "react-router";

export default function NotFoundPage() {
  return (
    <div className="flex flex-col items-center justify-center min-h-screen gap-4">
      <p className="text-6xl font-bold">404</p>
      <p className="text-sm font-semibold text-muted-foreground">Not Found</p>
      <Button render={<Link to="/" />} nativeButton={false}>Home</Button>
    </div>
  );
}
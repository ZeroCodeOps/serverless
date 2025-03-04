import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { logout, getUsername } from '../utils/auth';
import ThemeToggle from './ThemeToggle';
import { Server } from 'lucide-react';

const Header: React.FC = () => {
  const router = useRouter();
  
  const handleLogout = (): void => {
    logout();
    router.push('/');
  };

  return (
    <header className="sticky top-0 z-10 backdrop-blur-md bg-background/95 border-b border-black">
      <div className="container mx-auto flex justify-between items-center h-16 px-4">
        <Link href="/dashboard">
          <span className="text-xl font-semibold flex items-center gap-2 hover:text-primary transition-colors">
            <Server size={24} />
            Serverless Dashboard
          </span>
        </Link>
        <div className="flex items-center gap-4">
          <span className="text-sm text-muted-foreground">Welcome, {getUsername()}</span>
          <ThemeToggle />
          <button 
            onClick={handleLogout}
            className="btn btn-sm btn-destructive"
            aria-label="Logout"
          >
            Logout
          </button>
        </div>
      </div>
    </header>
  );
};

export default Header;
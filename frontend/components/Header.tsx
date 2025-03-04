import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { logout, getUsername } from '../utils/auth';

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
            <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" viewBox="0 0 24 24" fill="none" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 12h14M12 5l7 7-7 7" />
            </svg>
            Serverless Dashboard
          </span>
        </Link>
        <div className="flex items-center gap-4">
          <span className="text-sm text-muted-foreground">Welcome, {getUsername()}</span>
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
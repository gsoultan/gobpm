import { UnstyledButton, rem, ThemeIcon, Stack, Text, Group, Box } from '@mantine/core';
import {
  LayoutGrid,
  Building2,
  CheckSquare,
  ClipboardList,
  Network,
  Settings,
  LogOut,
  Sun,
  Moon,
  FolderGit2,
  Play,
  Zap,
  ChevronRight,
  Users,
  ShieldCheck,
  type LucideIcon,
} from 'lucide-react';
import classes from './Sidebar.module.css';
import { useAppStore } from '../store/useAppStore';
import { Link } from '@tanstack/react-router';

interface NavbarLinkProps {
  icon: LucideIcon;
  label: string;
  active?: boolean;
  expanded?: boolean;
  to?: string;
  onClick?(): void;
}

function NavbarLink({ icon: Icon, label, active, expanded, to, onClick }: NavbarLinkProps) {
  const content = (
    <Group gap="sm" wrap="nowrap">
      <Icon style={{ width: rem(22), height: rem(22) }} strokeWidth={1.5} />
      {expanded && (
        <Text size="sm" fw={active ? 700 : 500} className={classes.linkLabel}>
          {label}
        </Text>
      )}
    </Group>
  );

  if (to) {
    return (
      <Link 
        to={to}
        className={classes.link} 
        activeProps={{ 'data-active': true }}
        data-expanded={expanded || undefined}
        onClick={onClick}
      >
        {content}
        {!expanded && <div className={classes.activeIndicator} />}
      </Link>
    );
  }

  return (
    <UnstyledButton 
      onClick={onClick} 
      className={classes.link} 
      data-active={active || undefined}
      data-expanded={expanded || undefined}
    >
      {content}
      {active && !expanded && <div className={classes.activeIndicator} />}
    </UnstyledButton>
  );
}

const mainLinksData = [
  { icon: LayoutGrid, label: 'Dashboard', to: '/' },
  { icon: Building2, label: 'Organizations', to: '/organizations' },
  { icon: FolderGit2, label: 'Projects', to: '/projects' },
  { icon: ClipboardList, label: 'Inbox', to: '/inbox' },
  { icon: CheckSquare, label: 'Tasks', to: '/tasks' },
  { icon: Play, label: 'Instances', to: '/instances' },
  { icon: Network, label: 'Models', to: '/models' },
  { icon: Zap, label: 'Connectors', to: '/connectors' },
  { icon: Users, label: 'Users', to: '/users' },
  { icon: ShieldCheck, label: 'Groups', to: '/groups' },
];

export function Sidebar() {
  const { theme, toggleTheme, sidebarExpanded, toggleSidebar, clearAuth } = useAppStore();
  
  const links = mainLinksData.map((link) => (
    <NavbarLink
      {...link}
      key={link.label}
      expanded={sidebarExpanded}
    />
  ));

  return (
    <nav className={sidebarExpanded ? classes.navbarExpanded : classes.navbar}>
      <Box className={classes.header}>
        <Group justify={sidebarExpanded ? "space-between" : "center"} px={sidebarExpanded ? "md" : 0} wrap="nowrap">
          <ThemeIcon size="lg" radius="md" variant="gradient" gradient={{ from: 'blue', to: 'cyan', deg: 135 }}>
            <Network size={20} />
          </ThemeIcon>
          {sidebarExpanded && <Text fw={900} size="lg" variant="gradient" gradient={{ from: 'blue', to: 'cyan' }}>HERMOD</Text>}
          <UnstyledButton onClick={toggleSidebar} className={classes.expandButton}>
            <ChevronRight 
              size={16} 
              style={{ 
                transform: sidebarExpanded ? 'rotate(180deg)' : 'none',
                transition: 'transform 200ms ease'
              }} 
            />
          </UnstyledButton>
        </Group>
      </Box>
      
      <Stack gap={4} className={classes.navbarMain} px="sm">
        {links}
      </Stack>

      <Stack gap={4} className={classes.footer} px="sm">
        <NavbarLink 
          icon={theme === 'dark' ? Sun : Moon} 
          label={theme === 'dark' ? 'Light Mode' : 'Dark Mode'} 
          expanded={sidebarExpanded}
          onClick={toggleTheme}
        />
        <NavbarLink icon={Settings} label="Settings" expanded={sidebarExpanded} />
        <NavbarLink icon={LogOut} label="Logout" expanded={sidebarExpanded} onClick={clearAuth} />
      </Stack>
    </nav>
  );
}

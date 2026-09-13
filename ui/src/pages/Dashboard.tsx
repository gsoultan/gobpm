import { 
  Grid, 
  Card, 
  Text, 
  Group, 
  Stack, 
  ThemeIcon, 
  Box, 
  Title, 
  Button, 
  Badge, 
  rem, 
  Progress,
} from '@mantine/core';
import type { LucideIcon } from 'lucide-react';
import { 
  GitBranch, 
  TrendingUp, 
  Activity, 
  CheckCircle, 
  AlertCircle,
} from 'lucide-react';
import { 
  useDefinitions, 
  useProjects,
  useProcessStatistics,
  useInstances,
} from '../hooks/useProcess';
import { useAppStore } from '../store/useAppStore';
import { PageHeader } from '../components/PageHeader';
import { BusinessTimeline } from '../components/BusinessTimeline';
import { Link } from '@tanstack/react-router';
import { ComingSoonButton } from '../components/state/ComingSoon';
import { StatsLoadingState, ErrorState } from '../components/state';

/**
 * A single headline number.
 *
 * There is deliberately no `trend` prop. It used to accept a string, and every
 * call site passed a hardcoded one ("+12%", "+5%", "+2%") rendered beside a
 * green upward arrow and the words "vs last month" — while no endpoint in the
 * product computes a trend of any kind. Removing the prop means the fabrication
 * cannot come back without someone first building the data.
 *
 * `progress` is only for values that genuinely are a percentage of a whole.
 */
function StatCard({
  title,
  value,
  icon: Icon,
  color,
  progress,
  progressLabel,
  hint,
}: {
  title: string;
  value: React.ReactNode;
  icon: LucideIcon;
  color: string;
  progress?: number;
  progressLabel?: string;
  hint?: string;
}) {
  return (
    <Card shadow="md" radius="lg">
      <Group justify="space-between" align="flex-start" mb="sm">
        <Stack gap={0}>
          <Text size="xs" c="dimmed" fw={700} tt="uppercase" lts={rem(1)}>
            {title}
          </Text>
          <Text size="xl" fw={800} mt={5} style={{ fontSize: rem(28) }}>
            {value}
          </Text>
        </Stack>
        <ThemeIcon size={rem(48)} radius="md" variant="light" color={color}>
          <Icon size={rem(24)} />
        </ThemeIcon>
      </Group>
      
      {progress !== undefined && (
        <Stack gap={4} mt="md">
          <Group justify="space-between" align="flex-end">
            <Text size="xs" c="dimmed" fw={600}>{progressLabel ?? 'Complete'}</Text>
            <Text size="xs" fw={700} c={color}>{progress}%</Text>
          </Group>
          <Progress
            value={progress}
            color={color}
            size="sm"
            radius="xl"
            aria-label={`${progressLabel ?? 'Complete'}: ${progress}%`}
          />
        </Stack>
      )}

      {hint && (
        <Text size="xs" c="dimmed" mt="md">{hint}</Text>
      )}
    </Card>
  );
}

export function Dashboard() {
  const { currentProjectId, currentOrganizationId } = useAppStore();
  const { data: statsData, isLoading: statsLoading, error: statsError, refetch: refetchStats } = useProcessStatistics();
  const { data: defs } = useDefinitions();
  const { data: projectsData } = useProjects(currentOrganizationId);
  const { data: instancesData } = useInstances();
  

  // Falling back to zeros made an unloaded dashboard indistinguishable from a
  // real, idle one — "we don't know yet" rendered as "we know, and it's none".
  // The zeros remain only as a shape for the render below; statsLoading decides
  // whether they are ever shown.
  /*
   * These come from the Connect (protobuf) client, which serialises field
   * names in camelCase — activeInstances, not active_instances.
   *
   * Every read here used snake_case, so each one evaluated to `undefined` and
   * the `|| 0` fallback rendered a zero. The dashboard showed 0 active
   * instances, 0 process models and 0% completion no matter what the system
   * was actually doing, and nothing caught it because the processService
   * facade was typed `any`.
   */
  const stats = statsData?.stats;
  const activeInstances = stats?.activeInstances ?? 0;
  const failedInstances = stats?.failedInstances ?? 0;
  const totalTasks = stats?.totalTasks ?? 0;
  const pendingTasks = stats?.pendingTasks ?? 0;

  const lastInstanceId = instancesData?.instances?.[0]?.id;

  const totalDefinitions = defs?.definitions?.length || 0;
  const totalProjects = projectsData?.projects?.length || 0;


  if (!currentProjectId) {
    return (
      <Stack gap="xl">
        <PageHeader 
          title="Welcome to Metis BPM" 
          description="Get started by selecting or creating a project."
        />
        
        <Card shadow="sm" radius="lg" withBorder py={60}>
          <Stack align="center" gap="md">
            <ThemeIcon size={80} radius="xl" variant="light" color="indigo">
              <TrendingUp size={40} />
            </ThemeIcon>
            <Title order={2}>Ready to automate?</Title>
            {/*
              A project is now chosen automatically, so this is reached when
              there is none to choose rather than because somebody skipped a
              step. Telling them to "select one from the header" when the header
              is empty was the old, unhelpful half of this.
            */}
            <Text c="dimmed" ta="center" maw={500}>
              {totalProjects > 0
                ? 'Loading your project. If this stays here, pick one from the header.'
                : "Projects group related process models, tasks and instances. You'll need one to start."}
            </Text>

            {totalProjects === 0 && (
              <Button component={Link} to="/projects" size="md" radius="md" color="indigo">
                Create your first project
              </Button>
            )}
          </Stack>
        </Card>
      </Stack>
    );
  }

  const completionRate = totalTasks > 0 
    ? Math.round(((totalTasks - pendingTasks) / totalTasks) * 100) 
    : 0;

  return (
    <Stack gap="xl">
      <PageHeader 
        title="Dashboard" 
        description="Overview of your business processes and tasks."
        actions={
          <ComingSoonButton variant="light" leftSection={<Activity size={16} />} label="Report export is not implemented yet">
            Generate Report
          </ComingSoonButton>
        }
      />

      {statsLoading ? (
        <StatsLoadingState count={4} />
      ) : statsError ? (
        <ErrorState error={statsError} action="load your statistics" onRetry={() => refetchStats()} />
      ) : (
      <Grid gap="xl">
        <Grid.Col span={{ base: 12, md: 3 }}>
          <StatCard
            title="Active Instances"
            value={activeInstances}
            icon={Activity}
            color="indigo"
            hint="Processes currently running"
          />
        </Grid.Col>
        <Grid.Col span={{ base: 12, md: 3 }}>
          <StatCard
            title="Process Models"
            value={totalDefinitions}
            icon={GitBranch}
            color="teal"
            hint="Deployed definitions in this project"
          />
        </Grid.Col>
        <Grid.Col span={{ base: 12, md: 3 }}>
          <StatCard
            title="Tasks Completed"
            value={`${completionRate}%`}
            icon={CheckCircle}
            color="orange"
            progress={completionRate}
            progressLabel={`${totalTasks - pendingTasks} of ${totalTasks}`}
          />
        </Grid.Col>
        <Grid.Col span={{ base: 12, md: 3 }}>
          <StatCard
            title="Needs Attention"
            value={failedInstances}
            icon={AlertCircle}
            color={failedInstances > 0 ? 'red' : 'green'}
            hint={
              failedInstances > 0
                ? 'Failed instances waiting on someone'
                : 'Nothing has failed'
            }
          />
        </Grid.Col>
      </Grid>
      )}

      <Grid gap="xl">
        <Grid.Col span={12}>
          <Card shadow="sm" radius="lg" withBorder h="100%">
            <Group justify="space-between" mb="xl">
              <Group gap="sm">
                <Title order={4}>Business Timeline</Title>
                <Badge variant="light" color="indigo" radius="sm">Recent Activity</Badge>
              </Group>
              <Button component={Link} to="/instances" variant="subtle" size="xs">View all instances</Button>
            </Group>
            
            {lastInstanceId ? (
              <BusinessTimeline instanceId={lastInstanceId} />
            ) : (
              <Stack align="center" py={60} gap="sm">
                <ThemeIcon size={60} radius="xl" variant="light" color="gray">
                  <Activity size={32} />
                </ThemeIcon>
                <Text fw={700}>No recent activity</Text>
                <Text size="sm" c="dimmed">Start a process to see the activity timeline here.</Text>
              </Stack>
            )}
          </Card>
        </Grid.Col>
        
      </Grid>


      <Card shadow="sm" radius="lg" withBorder mb="xl">
        <Group justify="space-between" mb="xl">
          <Group gap="sm">
            <ThemeIcon color="yellow" variant="light" size="lg">
              <GitBranch size={20} />
            </ThemeIcon>
            <Title order={4}>Starter Templates</Title>
            <Badge variant="light" color="yellow">Recommended</Badge>
          </Group>
          <Text size="xs" c="dimmed">Quick start by picking a template</Text>
        </Group>
        
        <Grid gap="md">
          {[
            { 
              title: "Simple Approval", 
              desc: "A basic two-step approval process with conditional branching.", 
              color: "green",
              icon: CheckCircle
            },
            { 
              title: "Support Ticket", 
              desc: "Escalate issues based on priority and notify stakeholders.", 
              color: "orange",
              icon: AlertCircle
            },
            { 
              title: "Invoice Processing", 
              desc: "Automate invoice verification and payment triggering.", 
              color: "blue",
              icon: Activity
            }
          ].map(t => (
            <Grid.Col span={{ base: 12, md: 4 }} key={t.title}>
              <Card 
                withBorder 
                padding="md" 
                radius="md" 
                style={{ height: '100%' }}
              >
                 <Stack align="center" ta="center" gap="sm">
                   <ThemeIcon variant="light" color={t.color} size={50} radius="xl">
                      <t.icon size={24} />
                   </ThemeIcon>
                   <Box>
                      <Text size="md" fw={700}>{t.title}</Text>
                      <Text size="xs" c="dimmed" mt={4}>{t.desc}</Text>
                   </Box>
                   {/* Templates are not implemented; the card was a preview
                       with a button that did nothing when pressed. */}
                   <ComingSoonButton variant="light" color={t.color} size="xs" fullWidth mt="xs"
                                     label="Process templates are not available yet">
                      Use Template
                   </ComingSoonButton>
                 </Stack>
              </Card>
            </Grid.Col>
          ))}
        </Grid>
      </Card>
    </Stack>
  );
}

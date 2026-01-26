import { Container, Title, Text, Button, Group, Stack } from "@mantine/core";
import { IconCalendar, IconCheck } from "@tabler/icons-react";

export default function Home() {
  return (
    <Container size="md" py="xl">
      <Stack gap="xl" align="center" mt={100}>
        <IconCalendar size={64} stroke={1.5} />

        <Title order={1} ta="center">
          Hotdesk Booking System
        </Title>

        <Text size="lg" c="dimmed" ta="center" maw={600}>
          Modern hot-desking management platform for flexible office spaces.
          Book desks, track availability, and manage workspace resources efficiently.
        </Text>

        <Group>
          <Button
            size="lg"
            leftSection={<IconCheck size={20} />}
          >
            Get Started
          </Button>
          <Button
            size="lg"
            variant="outline"
          >
            View Docs
          </Button>
        </Group>

        <Stack gap="sm" mt="xl">
          <Text size="sm" c="dimmed">
            ✅ Next.js 16 with App Router
          </Text>
          <Text size="sm" c="dimmed">
            ✅ Mantine UI v7
          </Text>
          <Text size="sm" c="dimmed">
            ✅ TypeScript with Strict Mode
          </Text>
          <Text size="sm" c="dimmed">
            ✅ TanStack Query + Zod
          </Text>
        </Stack>
      </Stack>
    </Container>
  );
}

import '@mantine/core/styles.css';
import {
  AppShell,
  Group,
  Grid,
  Tabs,
  Text,
  Avatar,
} from '@mantine/core';
import { MantineProvider } from '@mantine/core';
import { ModalsProvider } from '@mantine/modals';
import { DevicesTable } from './Devices';
import { IconDeviceUnknown, IconSettings } from '@tabler/icons-react';
import { Footer } from './Footer';
import { Status } from './Status';
import './App.css';

function MyApp() {
  return (
 <AppShell
      padding="md"
      header={{ height: 100 }}
      footer={{ height: 60 }}
    >
      <AppShell.Header style={{backgroundColor: '#0ec28cff'}}>
        <Grid style={{height: '100%'}}>
        <Grid.Col span={6} style={{height: '100%'}}>
          <div style={{display: "block", padding: "var(--mantine-spacing-md)" }}>
            <Group wrap="nowrap">
              <Avatar src="/deltadore32.png" size={64} radius="md"/>
              <div  style={{ flex: 1 }}>
                <Text size="xl" fw={500}>
                  Go-Tydom
                </Text>
              </div>
            </Group>
          </div>
        </Grid.Col>
        <Grid.Col span={6} style={{height: '100%'}}>
          <Status/>
        </Grid.Col>
        </Grid>
      </AppShell.Header>
      <AppShell.Main>
        <Tabs defaultValue="devices">
        <Tabs.List>
          <Tabs.Tab value="devices" leftSection={<IconDeviceUnknown size={14} />}>
            Devices
          </Tabs.Tab>
          <Tabs.Tab value="settings" leftSection={<IconSettings size={14} />}>
            Settings
          </Tabs.Tab>
        </Tabs.List>
        
        <Tabs.Panel value="devices">
          <DevicesTable></DevicesTable>
        </Tabs.Panel>

        <Tabs.Panel value="settings">
            Settings
        </Tabs.Panel>
        </Tabs>
      </AppShell.Main>
      <AppShell.Footer><Footer /></AppShell.Footer>
    </AppShell>
  )
}

export default function App() {
  return <MantineProvider><ModalsProvider>{MyApp()}</ModalsProvider></MantineProvider>;
}
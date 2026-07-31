import { IconBrandGithub, IconHome } from '@tabler/icons-react';
import { ActionIcon, Grid, Group, Text } from '@mantine/core';
import { useState, useEffect } from 'react';
import { MyTydomService } from './tydom';

interface Infos {
  version: string
}

export function Footer() {
  const tydomClient = new MyTydomService()
  const [appInfo, setAppInfos] = useState<Infos>();
  useEffect(() => {
    const fetchData = async () => {
      try {
        const infos = await tydomClient.getInfos();
        setAppInfos(infos.data)
      } catch (err: any) {
        setAppInfos({version: "unknown"})
      }
    };
    fetchData();
  }, []); // Empty dependency array ensures it runs once on mount

  const openInNewTab = (url: string) => {
    window.open(url, '_blank', 'noopener,noreferrer');
  };

  return (
    <div>
    <Grid style={{height: '100%'}}>
    <Grid.Col span={6} style={{height: '100%'}}>
      <div style={{display: "block", padding: "var(--mantine-spacing-md)" }}>
        <Text c="dimmed" size="sm">
          Version {appInfo?.version}
        </Text>
      </div>
    </Grid.Col>
    <Grid.Col span={6} style={{height: '100%'}}>
      <div style={{display: "block", padding: "var(--mantine-spacing-md)" }}>
        <Group justify="right" wrap="nowrap">
          <ActionIcon size="lg" color="gray" variant="subtle" onClick={() => openInNewTab('https://github.com/dchauviere/go-tydom')}>
            <IconBrandGithub size={18} stroke={1.5} />
          </ActionIcon>
          <ActionIcon size="lg" color="gray" variant="subtle" onClick={() => openInNewTab('https://home-assistant.io')}>
            <IconHome size={18} stroke={1.5} />
          </ActionIcon>
        </Group>
      </div>
    </Grid.Col>
    </Grid>
    </div>
  );
}
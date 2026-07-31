import { Indicator, Image, Group } from '@mantine/core';
import { useState, useEffect } from 'react';
import { MyTydomService } from './tydom';

interface Status {
  mqtt: string
  tydom: string
}

export function Status() {
  const tydomClient = new MyTydomService()
  const [appStatus, setAppStatus] = useState<Status>();
  useEffect(() => {
    const fetchData = async () => {
      try {
        const status = await tydomClient.getStatus();
        let finalStatus = {mqtt: "grey", tydom: "grey"}
        if (status.data.mqtt) {
            finalStatus.mqtt = "green"
        } else {
            finalStatus.mqtt = "red"
        }
        if (status.data.tydom) {
            finalStatus.tydom = "green"
        } else {
            finalStatus.tydom = "red"
        }
        setAppStatus(finalStatus)
      } catch (err: any) {
        setAppStatus({mqtt: "grey", tydom: "grey"})
      }
    };
    let timerId = setTimeout(fetchData, 10000);
    return ()=> clearInterval(timerId); 
  });

  return (
    <div style={{display: "block", padding: "var(--mantine-spacing-md)" }}>
        <Group wrap="nowrap" justify='right' align="center">
          <Indicator inline size={8} color={appStatus?.mqtt}>
            <Image radius="md" src="/mqtt32.png" h={32} w="auto"/>
          </Indicator>

          <Indicator inline size={8} color={appStatus?.tydom}>
            <Image radius="md" src="/tydomgw32.png" h={32} w="auto"/>
          </Indicator>
        </Group>
    </div>
  );
}
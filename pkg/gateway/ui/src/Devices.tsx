import { useState, useEffect } from 'react';
import { Modal, TextInput, Button, Table, ActionIcon } from '@mantine/core';
import { IconAdjustments, IconTrash } from '@tabler/icons-react';
import { MyTydomService, type Device } from './tydom';
import { modals } from '@mantine/modals';

export function DevicesTable() {
  const [devices, setDevices] = useState<Device[]>([]);
  const [modalOpen, setModalOpen] = useState(false);
  const [editingRow, setEditingRow] = useState<Device|null>(null);
  const [loading, setLoading] = useState(true); // State to manage loading
  const [error, setError] = useState(null); // State to handle errors

  const tydomClient = new MyTydomService()

  const openDeleteConfirm = (deleteRow: Device) => 
    modals.openConfirmModal({ 
      title: 'Delete device',
      centered: true,
      children: ( 
        <p>Realy delete device ?</p>
      ), 
      labels: { confirm: 'Delete', cancel: 'Cancel' },
      confirmProps: { color: 'red' }, 
      
      onConfirm: () => { 
        tydomClient.deleteDevice(deleteRow?.deviceId!)
        setDevices((prevData) => 
          prevData.filter((row) => 
            row.deviceId !== deleteRow?.deviceId
          )
        );
      },
    });
  
  useEffect(() => {
    const fetchData = async () => {
      try {
        setLoading(true); // Start loading
        const response = await tydomClient.getDevices();
        setDevices(response.data); // Save API response to state
      } catch (err: any) {
        setError(err.message); // Handle errors
      } finally {
        setLoading(false); // Stop loading
      }
    };
    fetchData();
  }, []); // Empty dependency array ensures it runs once on mount

  if (loading) return <p>Loading...</p>;
  if (error) return <p>Error: {error}</p>;

  const handleDelete = (row: Device) => {
    setEditingRow(row);
    openDeleteConfirm(row);
  }

  const handleEdit = (row: Device) => {
    setEditingRow(row);
    setModalOpen(true);
  };

  const handleSave = () => {
    if (editingRow) {
      tydomClient.setDeviceName(editingRow?.deviceId, editingRow?.endpointId, editingRow?.name)
      setDevices((prevData) =>
        prevData.map((row) =>
          row.deviceId === editingRow.deviceId && row.endpointId === editingRow.endpointId && row.name !== editingRow.name ? {...row, ...editingRow} : row
        )
      );
    };
    setModalOpen(false);
  };

  return (
    <>
      <Table>
        <Table.Thead>
          <Table.Tr>
            <Table.Th>Name</Table.Th>
            <Table.Th>Type</Table.Th>
            <Table.Th></Table.Th>
          </Table.Tr>
        </Table.Thead>
        <Table.Tbody>
        { devices.map((element: Device) => (
          <Table.Tr key={element.deviceId}>
            <Table.Td align='left'>{element.name}</Table.Td>
            <Table.Td align='left'>{element.type}</Table.Td>
            <Table.Td>
              <ActionIcon.Group>
                <ActionIcon variant="filled" aria-label="Settings" onClick={() => handleEdit(element)}>
                  <IconAdjustments style={{ width: '70%', height: '70%' }} stroke={1.5} />
                </ActionIcon>
                <ActionIcon variant="filled" color="red" aria-label="Delete" onClick={() => handleDelete(element)}>
                  <IconTrash style={{ width: '70%', height: '70%' }} stroke={1.5} />
                </ActionIcon>
              </ActionIcon.Group>
            </Table.Td>
          </Table.Tr>))
        }
        </Table.Tbody>
      </Table>
      
      <Modal
        opened={modalOpen}
        onClose={() => setModalOpen(false)}
        title="Edit Row"
      >
        {editingRow && (
          <>
            <TextInput
              label="Name"
              value={editingRow.name}
              onChange={(e) =>
                setEditingRow({ ...editingRow, name: e.target.value })
              }
            />
            <Button onClick={handleSave} mt="md">
              Save
            </Button>
          </>
        )}
      </Modal>
    </>
);
}
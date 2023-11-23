import { useState } from 'react';
import { TextInput, TextInputProps, ActionIcon, Button, Paper, Text, Stack, Container, Textarea, Skeleton, SimpleGrid, Grid, useMantineTheme, rem } from '@mantine/core';
import { IconSearch, IconArrowRight } from '@tabler/icons-react';

type Message = {
    text: string;
    user: string;
};

const child = <Skeleton height={140} radius="md" animate={false} />;
const PRIMARY_COL_HEIGHT = rem(500);



export default function Chat() {
    const [messages, setMessages] = useState<Message[]>([]);
    const [newMessage, setNewMessage] = useState('');

    const handleSend = () => {
        if (newMessage.trim()) {
            setMessages([...messages, { text: newMessage, user: 'You' }]);
            setNewMessage('');
        }
    };
    const theme = useMantineTheme();

    return (
        <Container style={{ padding: '5rem', width: '100vw', height: '100vh' }}>
            <Grid>
                <Grid.Col span={8}>
                    <Paper style={{ height: PRIMARY_COL_HEIGHT, overflowY: 'scroll', scrollbarWidth: 'none' }}>
                        <Stack style={{ overflowY: 'scroll', flexGrow: 1 }}>
                            {messages.map((message, index) => (
                                <Paper
                                    key={index}
                                    shadow="xs"
                                    p="md"
                                    style={{
                                        maxWidth: '70%',
                                        margin: '10px',
                                        backgroundColor: message.user === 'You' ? '#blue' : '#grey',
                                    }}
                                >
                                    <Text style={{ marginBottom: '5px', fontWeight: 500 }}>
                                        {message.user}
                                    </Text>
                                    <Text size="sm">{message.text}</Text>
                                </Paper>
                            ))}
                        </Stack>
                    </Paper>
                    <Paper style={{ marginTop: 'md' }}>
                        <Textarea
                            placeholder="Type your message"
                            value={newMessage}
                            onChange={(event) => setNewMessage(event.currentTarget.value)}
                            minRows={1}
                            autosize
                            radius="lg"
                            rightSectionWidth={90}
                            rightSection={
                                <Button radius="lg" color={theme.primaryColor} variant="filled"
                                // style={{ paddingLeft: rem(5), paddingRight: rem(5)}}
                                >
                                    Send
                                    {/* <IconArrowRight style={{ paddingRight: rem(5), width: rem(28), height: rem(18) }} stroke={1.5} /> */}
                                </Button>
                            }
                            autoFocus
                        />
                    </Paper>
                </Grid.Col>
                <Grid.Col span={4} style={{ width: '20%' }}>
                    <Skeleton height={PRIMARY_COL_HEIGHT} radius="md" animate={false} />
                </Grid.Col>
            </Grid>
        </Container>
    );

    // dissabledc 
    return (
        <Container style={{ display: 'flex', flexDirection: 'column', height: '100vh' }}>


            <Stack style={{ overflowY: 'scroll', flexGrow: 1 }}>
                {messages.map((message, index) => (
                    <Paper
                        key={index}
                        shadow="xs"
                        p="md"
                        style={{
                            maxWidth: '70%',
                            margin: '10px',
                            backgroundColor: message.user === 'You' ? '#blue' : '#grey',
                        }}
                    >
                        <Text style={{ marginBottom: '5px', fontWeight: 500 }}>
                            {message.user}
                        </Text>
                        <Text size="sm">{message.text}</Text>
                    </Paper>
                ))}
            </Stack>
            <Stack style={{ position: 'fixed', bottom: 0, left: 0, right: 0, padding: '10px' }}>
                <Textarea
                    placeholder="Type your message"
                    value={newMessage}
                    onChange={(event) => setNewMessage(event.currentTarget.value)}
                    style={{ marginBottom: '10px' }}
                />
                <Button onClick={handleSend}>Send</Button>
            </Stack>
        </Container>
    );
}

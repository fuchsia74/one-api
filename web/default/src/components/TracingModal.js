import React, { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import {
  Modal,
  Header,
  Segment,
  Grid,
  Label,
  Icon,
  Loader,
  Message,
  Divider,
  Progress,
  Table,
  Button,
} from 'semantic-ui-react';
import { API, showError, timestamp2string } from '../helpers';

const TracingModal = ({ open, onClose, logId }) => {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [traceData, setTraceData] = useState(null);
  const [error, setError] = useState(null);

  useEffect(() => {
    if (open && logId) {
      fetchTraceData();
    }
  }, [open, logId]);

  const fetchTraceData = async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await API.get(`/api/trace/log/${logId}`);
      if (res.data.success) {
        setTraceData(res.data.data);
      } else {
        setError(res.data.message || 'Failed to fetch trace data');
      }
    } catch (err) {
      setError(err.response?.data?.message || 'Failed to fetch trace data');
    } finally {
      setLoading(false);
    }
  };

  const formatDuration = (milliseconds) => {
    if (!milliseconds) return 'N/A';
    if (milliseconds < 1000) {
      return `${milliseconds}ms`;
    }
    return `${(milliseconds / 1000).toFixed(2)}s`;
  };

  const formatTimestamp = (timestamp) => {
    if (!timestamp) return 'N/A';
    return timestamp2string(Math.floor(timestamp / 1000));
  };

  const getStatusColor = (status) => {
    if (status >= 200 && status < 300) return 'green';
    if (status >= 300 && status < 400) return 'yellow';
    if (status >= 400 && status < 500) return 'orange';
    if (status >= 500) return 'red';
    return 'grey';
  };

  const renderTimeline = () => {
    if (!traceData?.timestamps) return null;

    const { timestamps, durations } = traceData;
    const timelineEvents = [];

    if (timestamps.request_received) {
      timelineEvents.push({
        key: 'received',
        title: 'Request Received',
        timestamp: timestamps.request_received,
        icon: 'play',
        color: 'blue',
      });
    }

    if (timestamps.request_forwarded) {
      timelineEvents.push({
        key: 'forwarded',
        title: 'Forwarded to Upstream',
        timestamp: timestamps.request_forwarded,
        icon: 'arrow right',
        color: 'teal',
        duration: durations?.processing_time,
      });
    }

    if (timestamps.first_upstream_response) {
      timelineEvents.push({
        key: 'upstream_response',
        title: 'First Upstream Response',
        timestamp: timestamps.first_upstream_response,
        icon: 'reply',
        color: 'purple',
        duration: durations?.upstream_response_time,
      });
    }

    if (timestamps.first_client_response) {
      timelineEvents.push({
        key: 'client_response',
        title: 'First Client Response',
        timestamp: timestamps.first_client_response,
        icon: 'send',
        color: 'orange',
        duration: durations?.response_processing_time,
      });
    }

    if (timestamps.upstream_completed) {
      timelineEvents.push({
        key: 'upstream_completed',
        title: 'Upstream Completed',
        timestamp: timestamps.upstream_completed,
        icon: 'check circle',
        color: 'green',
        duration: durations?.streaming_time,
      });
    }

    if (timestamps.request_completed) {
      timelineEvents.push({
        key: 'completed',
        title: 'Request Completed',
        timestamp: timestamps.request_completed,
        icon: 'flag checkered',
        color: 'green',
      });
    }

    return (
      <Segment>
        <Header as="h4">
          <Icon name="clock" />
          Request Timeline
        </Header>
        <Table basic="very" celled>
          <Table.Header>
            <Table.Row>
              <Table.HeaderCell>Event</Table.HeaderCell>
              <Table.HeaderCell>Timestamp</Table.HeaderCell>
              <Table.HeaderCell>Duration</Table.HeaderCell>
            </Table.Row>
          </Table.Header>
          <Table.Body>
            {timelineEvents.map((event, index) => (
              <Table.Row key={event.key}>
                <Table.Cell>
                  <Label color={event.color}>
                    <Icon name={event.icon} />
                    {event.title}
                  </Label>
                </Table.Cell>
                <Table.Cell>{formatTimestamp(event.timestamp)}</Table.Cell>
                <Table.Cell>
                  {event.duration ? formatDuration(event.duration) : '-'}
                </Table.Cell>
              </Table.Row>
            ))}
          </Table.Body>
        </Table>

        {durations?.total_time && (
          <Segment color="blue">
            <Header as="h5">
              <Icon name="stopwatch" />
              Total Request Time: {formatDuration(durations.total_time)}
            </Header>
          </Segment>
        )}
      </Segment>
    );
  };

  const renderRequestInfo = () => {
    if (!traceData) return null;

    return (
      <Segment>
        <Header as="h4">
          <Icon name="info circle" />
          Request Information
        </Header>
        <Grid columns={2} divided>
          <Grid.Row>
            <Grid.Column>
              <Label>
                <Icon name="linkify" />
                URL
              </Label>
              <div style={{ marginTop: '5px', wordBreak: 'break-all' }}>
                {traceData.url}
              </div>
            </Grid.Column>
            <Grid.Column>
              <Label>
                <Icon name="code" />
                Method
              </Label>
              <div style={{ marginTop: '5px' }}>
                <Label color="blue">{traceData.method}</Label>
              </div>
            </Grid.Column>
          </Grid.Row>
          <Grid.Row>
            <Grid.Column>
              <Label>
                <Icon name="file" />
                Body Size
              </Label>
              <div style={{ marginTop: '5px' }}>
                {traceData.body_size ? `${traceData.body_size} bytes` : 'N/A'}
              </div>
            </Grid.Column>
            <Grid.Column>
              <Label>
                <Icon name="flag" />
                Status
              </Label>
              <div style={{ marginTop: '5px' }}>
                <Label color={getStatusColor(traceData.status)}>
                  {traceData.status || 'N/A'}
                </Label>
              </div>
            </Grid.Column>
          </Grid.Row>
          <Grid.Row>
            <Grid.Column>
              <Label>
                <Icon name="tag" />
                Trace ID
              </Label>
              <div style={{ marginTop: '5px', fontFamily: 'monospace', fontSize: '0.9em' }}>
                {traceData.trace_id}
              </div>
            </Grid.Column>
            <Grid.Column>
              <Label>
                <Icon name="user" />
                User
              </Label>
              <div style={{ marginTop: '5px' }}>
                {traceData.log?.username || 'N/A'}
              </div>
            </Grid.Column>
          </Grid.Row>
        </Grid>
      </Segment>
    );
  };

  return (
    <Modal open={open} onClose={onClose} size="large">
      <Modal.Header>
        <Icon name="chart line" />
        Request Tracing Details
      </Modal.Header>
      <Modal.Content scrolling>
        {loading && (
          <Segment>
            <Loader active inline="centered">
              Loading trace data...
            </Loader>
          </Segment>
        )}

        {error && (
          <Message negative>
            <Message.Header>Error</Message.Header>
            <p>{error}</p>
          </Message>
        )}

        {traceData && !loading && (
          <>
            {renderRequestInfo()}
            <Divider />
            {renderTimeline()}
          </>
        )}
      </Modal.Content>
      <Modal.Actions>
        <Button onClick={onClose}>
          <Icon name="close" />
          Close
        </Button>
      </Modal.Actions>
    </Modal>
  );
};

export default TracingModal;

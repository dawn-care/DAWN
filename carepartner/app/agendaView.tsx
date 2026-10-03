/**
 * AgendaView.tsx
 * 
 * User: Carepartner
 * 
 * Displays and manages playback of the daily scheduled routine.
 * 
 * Responsibilities:
 * - Build a playback queue from today's schedule
 * - Render appropriate media/content screens (photo, audio, video, widgets, text)
 * - Handle timed vs non-timed item execution
 * - Track and skip already played items
 * - Auto-advance content based on duration or playback completion
 * 
 * Navigation:
 * - On completion → redirects to `/`
 * - On error or empty state → redirects to `/`
 * 
 * Notes:
 * - Queue is built once per session to prevent replay loops
 * - Uses ScheduleContext for data, playback tracking, and refresh
 */

import { useState, useEffect, useRef } from 'react';
import TextViewerScreen from './text';
import WidgetViewerScreen from './widget';
import PhotoViewerScreen from './photo';
import { useSchedule } from './ScheduleContext';
import { API_URL } from '../src/api/config';
import { resolveApiUri } from '../src/api/resolveApiUri';
import Player from '../src/playback/Player';
import React from 'react';
import { Pressable, SafeAreaView, StyleSheet, Text, View } from 'react-native';
import { router, useLocalSearchParams } from 'expo-router';

/* ============================
   Types
============================ */

/**
 * Metadata shapes returned inline by the backend
 */
export type PhotoMetadata = { id: number; uri: string; displayed_name: string; filename: string; extension: string };
export type VideoMetadata = { id: number; uri: string; displayed_name: string; filename: string; extension: string };
/** cover_uri: the song's embedded album art, a full URL or a path relative to the API. Not sent by the server yet. */
export type AudioMetadata = { id: number; uri: string; displayed_name: string; filename: string; extension: string; cover_uri?: string | null };
export type WeatherMetadata = { ID: number; ZIPCODE: string };
export type NewsMetadata = { ID: number; CATEGORY: string; LANGUAGE: string; COUNTRY: string };
export type MessageMetadata = { id: number; message: string; image_uri?: string | null };
export type EventMetadata = { id: number; list_id: number; time: string; text: string };
export type CarepartnerLocationMetadata = { id: number; location: string; message: string; image_uri: string };
export type CaregiverLocationMetadata = { id: number; location: string; message: string; image_uri: string };

/**
 * Matches the ConfigurationWithMetadata Go struct 
 * returned by GetConfigurationsByDayAndCarepartner
 */
export type ConfigurationWithMetadata = {
    id: number;
    caregiver_id: number;
    carepartner_id: number;
    module_id: number;
    occurence_type: string;
    occurence_definition: string[];
    position: number;
    duration: number;
    timed: boolean;
    time: string | null;
    module_type: string;
    metadata_id: number;
    metadata: PhotoMetadata | VideoMetadata | AudioMetadata | WeatherMetadata
        | NewsMetadata | MessageMetadata | EventMetadata
        | CarepartnerLocationMetadata | CaregiverLocationMetadata | null;
};

/* ============================
   Helpers
============================ */

/**
 * Reusable glass-style container for UI.
 */
function GlassCard({ children, style }: { children: React.ReactNode; style?: any }) {
    return <View style={[styles.glassBase, style]}>{children}</View>;
}

/**
 * Determines the header title based on module type + metadata.
 */
function getHeaderTitle(item: ConfigurationWithMetadata): string {
    if (item.module_type === 'audio') {
        const metadata = item.metadata as AudioMetadata | null;
        return metadata?.displayed_name ?? 'Audio';
    }

    if (item.module_type === 'video') {
        const metadata = item.metadata as VideoMetadata | null;
        return metadata?.displayed_name ?? 'Video';
    }

    if (item.module_type === 'photo') {
        const metadata = item.metadata as PhotoMetadata | null;
        return metadata?.displayed_name ?? 'Photos';
    }

    if (item.module_type === 'messages') {
        return 'Message';
    }

    if (item.module_type === 'weather') {
        return 'Weather';
    }

    if (item.module_type === 'sports') {
        return 'Sports News';
    }

    if (item.module_type === 'news') {
        return 'Good News';
    }

    const metadata = item.metadata as EventMetadata | null;
    return metadata?.time ?? 'Event';
}

/* ============================
   Component
============================ */

/**
 * AgendaView Component
 * 
 * Handles:
 *  - Building playback queue
 *  - Managing playback state
 *  - Rendering media components
 *  - Advancing through content
 */
export default function AgendaView(){

    /* ============================
        State
    ============================ */

    const { todaysItems, loading, error, refresh, playedIds, markPlayed } = useSchedule();
    const { trigger } = useLocalSearchParams<{ trigger?: string }>();

    /** Local playback state */
    const [playQueue, setPlayQueue] = useState<ConfigurationWithMetadata[]>([]);
    const [currentIndex, setCurrentIndex] = useState(0);
    const playQueueRef = useRef<ConfigurationWithMetadata[]>([]);

    /* ============================
        Effects
    ============================ */

    /** Keep ref in sync for use inside duration timer */
    useEffect(() => { playQueueRef.current = playQueue; }, [playQueue]);

    /** Refresh data every time agendaView mounts */
    useEffect(() => {
        refresh();
    }, []);

    /**
     * Build queue ONCE when data first loads. Played items are excluded.
     * The queue is NOT rebuilt on subsequent refreshes to prevent periodic replays.
     */
    const queueBuiltRef = useRef(false);
    useEffect(() => {
        if (queueBuiltRef.current) return;
        if (loading || !todaysItems.length) return;

        let queue: ConfigurationWithMetadata[];

        if (trigger === 'scheduled') {
            const currentTime = (() => {
                const now = new Date();
                return `${String(now.getHours()).padStart(2, '0')}:${String(now.getMinutes()).padStart(2, '0')}`;
            })();

            queue = todaysItems
                .filter(i => i.timed && (i.time ?? '').slice(0, 5) <= currentTime && !playedIds.has(i.id))
                .sort((a, b) => (a.time ?? '').localeCompare(b.time ?? ''));
        } else {
            queue = todaysItems.filter(i => !i.timed);
        }

        queueBuiltRef.current = true;
        setCurrentIndex(0);
        setPlayQueue(queue);
    }, [todaysItems, loading, trigger, playedIds]);

    /** When an error occurs, redirect to home after 5 seconds */
    useEffect(() => {
        if (!error) return;
        const timer = setTimeout(() => {
            router.replace('/');
        }, 5000);
        return () => clearTimeout(timer);
    }, [error]);

    /** When no items are available, redirect to home after 5 seconds */
    useEffect(() => {
        if (loading || error || playQueue.length > 0) return;
        const timer = setTimeout(() => {
            router.replace('/');
        }, 5000);
        return () => clearTimeout(timer);
    }, [loading, error, playQueue]);

    /** 
     * Auto-advance timer — re-runs when currentIndex changes OR when the queue is first populated
     * non-audio/video items
     */
    useEffect(() => {
        const queue = playQueueRef.current;
        if (!queue.length) return;

        const item = queue[currentIndex];
        if (!item) return;
        console.log('Displaying item:', item);

        if (item.module_type === 'audio' || item.module_type === 'video') {
            return;
        }

        const timer = setTimeout(() => {
            advanceToNext();
        }, (item.duration || 10) * 1000);

        return () => clearTimeout(timer);
    }, [currentIndex, playQueue]);

    /** Move to next item */
    const advanceToNext = () => {
        setCurrentIndex(prev => {
            const queue = playQueueRef.current;
            const current = queue[prev];
            if (current) markPlayed(current.id);
            if (prev + 1 >= queue.length) {
                router.replace('/');
                return prev;
            }
            return prev + 1;
        });
    };

    /** Navigation helpers */
    const goToPreviousItem = () => {
        setCurrentIndex((prev) => Math.max(prev - 1, 0));
    };
    const goToNextItem = () => {
        advanceToNext();
    };

    /** Skip media that can't be played, so one bad file doesn't hold up the queue */
    const skipFailedItem = (reason: string) => {
        console.error('Playback failed:', reason);
        advanceToNext();
    };

    const currentItem = playQueue[currentIndex];

    /* ============================
       UI STATES
    ============================ */

    if (loading) {
        return (
            <SafeAreaView style={styles.safe}>
                <View style={styles.errorContainer}>
                    <Text style={styles.errorTitle}>Loading…</Text>
                </View>
            </SafeAreaView>
        );
    }

    if (error) {
        return (
            <SafeAreaView style={styles.safe}>
                <View style={styles.errorContainer}>
                    <Text style={styles.errorTitle}>Unable to load agenda</Text>
                    <Text style={styles.errorMessage}>{error}</Text>
                    <Text style={styles.errorSubtext}>Returning to home screen…</Text>
                </View>
            </SafeAreaView>
        );
    }

    if (!currentItem) {
        return (
            <SafeAreaView style={styles.safe}>
                <View style={styles.errorContainer}>
                    <Text style={styles.errorTitle}>No items for today</Text>
                    <Text style={styles.errorSubtext}>Returning to home screen…</Text>
                </View>
            </SafeAreaView>
        );
    }

    const canGoPrev = currentIndex > 0;
    const canGoNext = playQueue.length > 0;
    const headerTitle = getHeaderTitle(currentItem);

    /* ============================
       RENDER
    ============================ */
    return (
        <SafeAreaView style={styles.safe}>
            <View style={styles.topBar}>
                <Text style={styles.topTitle}>{headerTitle}</Text>
            </View>

            <View style={styles.container}>
                {/* Render correct module type */}
                {(currentItem.module_type === "photo" || currentItem.module_type === "messages") && (
                    <PhotoViewerScreen configuration={currentItem} />
                )}

                {(currentItem.module_type === "weather" ||
                currentItem.module_type === "sports" ||
                currentItem.module_type === "news") && (
                    <WidgetViewerScreen configuration={currentItem} />
                )}

                {(currentItem.module_type === "event" || currentItem.module_type === "messages") && (
                    <TextViewerScreen configuration={currentItem} />
                )}

                {/* Keyed by item so each one gets a fresh player */}
                {(currentItem.module_type === "video" || currentItem.module_type === "audio") &&
                    currentItem.metadata && (
                    <Player
                        key={currentItem.id}
                        kind={currentItem.module_type}
                        uri={`${API_URL}/${currentItem.module_type}/file?id=${(currentItem.metadata as VideoMetadata | AudioMetadata).id}`}
                        title={(currentItem.metadata as VideoMetadata | AudioMetadata).displayed_name}
                        imageUri={resolveApiUri((currentItem.metadata as AudioMetadata).cover_uri)}
                        onFinished={goToNextItem}
                        onFailed={skipFailedItem}
                    />
                )}

                {/* Navigation arrows */}
                <Pressable
                    onPress={goToPreviousItem}
                    disabled={!canGoPrev}
                    style={({ pressed }) => [
                        styles.leftWrapper,
                        pressed && canGoPrev && styles.arrowPressed,
                        !canGoPrev && styles.arrowDisabled,
                    ]}
                >
                    <GlassCard style={styles.arrowGlass}>
                        <Text style={styles.arrowText}>‹</Text>
                    </GlassCard>
                </Pressable>

                <Pressable
                    onPress={goToNextItem}
                    disabled={!canGoNext}
                    style={({ pressed }) => [
                        styles.rightWrapper,
                        pressed && canGoNext && styles.arrowPressed,
                        !canGoNext && styles.arrowDisabled,
                    ]}
                >
                    <GlassCard style={styles.arrowGlass}>
                        <Text style={styles.arrowText}>›</Text>
                    </GlassCard>
                </Pressable>
            </View>
        </SafeAreaView>
    );
}

const styles = StyleSheet.create({
    safe: {
        flex: 1,
        backgroundColor: '#000',
    },
    topBar: {
        backgroundColor: '#c7f3f6',
        borderBottomWidth: 2,
        borderBottomColor: 'rgba(0,0,0,0.6)',
        paddingHorizontal: 16,
        paddingVertical: 14,
        alignItems: 'center',
        justifyContent: 'center',
    },
    topTitle: {
        fontSize: 54,
        fontWeight: '300',
        color: 'rgba(0,0,0,0.85)',
        textAlign: 'center',
    },
    container: {
        flex: 1,
        position: 'relative',
    },
    glassBase: {
        borderRadius: 16,
        borderWidth: 1.5,
        borderColor: 'rgba(0,0,0,0.25)',
        backgroundColor: 'rgba(255,255,255,0.55)',
    },
    leftWrapper: {
        position: 'absolute',
        left: 26,
        top: '50%',
        marginTop: -41,
    },
    rightWrapper: {
        position: 'absolute',
        right: 26,
        top: '50%',
        marginTop: -41,
    },
    arrowGlass: {
        width: 82,
        height: 82,
        borderRadius: 18,
        borderWidth: 2,
        borderColor: 'rgba(0, 255, 240, 0.9)',
        alignItems: 'center',
        justifyContent: 'center',
    },
    arrowText: {
        fontSize: 54,
        color: 'rgba(0,0,0,0.75)',
        lineHeight: 54,
        marginTop: -2,
    },
    arrowPressed: {
        transform: [{ scale: 0.98 }],
    },
    arrowDisabled: {
        opacity: 0.45,
    },
    errorContainer: {
        flex: 1,
        alignItems: 'center',
        justifyContent: 'center',
        padding: 32,
    },
    errorTitle: {
        fontSize: 36,
        fontWeight: '600',
        color: '#fff',
        marginBottom: 12,
        textAlign: 'center',
    },
    errorMessage: {
        fontSize: 20,
        color: 'rgba(255,255,255,0.7)',
        marginBottom: 24,
        textAlign: 'center',
    },
    errorSubtext: {
        fontSize: 18,
        color: 'rgba(255,255,255,0.45)',
        textAlign: 'center',
    },
});
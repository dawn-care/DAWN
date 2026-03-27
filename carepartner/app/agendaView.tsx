import { useState, useEffect, useRef } from 'react';
import AudioScreen from './audio';
import GreetingScreen from './greeting';
import TextViewerScreen from './text';
import PhotoViewerScreen from './photo';
import { useSchedule } from './ScheduleContext';
import React from 'react';
import { Pressable, SafeAreaView, StyleSheet, Text, View } from 'react-native';
import { router } from 'expo-router';

// Metadata shapes returned inline by the backend
export type PhotoMetadata = { id: number; uri: string; displayed_name: string; filename: string; extension: string };
export type VideoMetadata = { id: number; uri: string; displayed_name: string; filename: string; extension: string };
export type AudioMetadata = { id: number; uri: string; displayed_name: string; filename: string; extension: string };
export type WeatherMetadata = { ID: number; ZIPCODE: string };
export type NewsMetadata = { ID: number; CATEGORY: string; LANGUAGE: string; COUNTRY: string };
export type MessageMetadata = { id: number; message: string; image_uri: string };
export type EventMetadata = { id: number; list_id: number; time: string; text: string };
export type CarepartnerLocationMetadata = { id: number; location: string; message: string; image_uri: string };
export type CaregiverLocationMetadata = { id: number; location: string; message: string; image_uri: string };

// Matches the ConfigurationWithMetadata Go struct returned by GetConfigurationsByDayAndCarepartner
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

export type MediaPlaybackProps = {
    onPlaybackComplete: () => void;
};

function GlassCard({ children, style }: { children: React.ReactNode; style?: any }) {
    return <View style={[styles.glassBase, style]}>{children}</View>;
}

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

    if (item.module_type === 'message') {
        const metadata = item.metadata as MessageMetadata | null;
        return metadata?.message ?? 'Photos';
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

export default function AgendaView(){
    const { todaysItems, loading, error, refresh } = useSchedule();

    // --- Local playback state (only regular/non-timed items) ---
    const [playQueue, setPlayQueue] = useState<ConfigurationWithMetadata[]>([]);
    const [currentIndex, setCurrentIndex] = useState(0);
    const playQueueRef = useRef<ConfigurationWithMetadata[]>([]);

    // Keep ref in sync for use inside duration timer
    useEffect(() => { playQueueRef.current = playQueue; }, [playQueue]);

    // Refresh data every time agendaView mounts
    useEffect(() => {
        refresh();
    }, []);

    // Build queue from non-timed items only
    useEffect(() => {
        if (!todaysItems.length) return;
        const regular = todaysItems.filter(i => !i.timed);
        setCurrentIndex(0);
        setPlayQueue(regular);
    }, [todaysItems]);

    // When an error occurs, redirect to home after 5 seconds
    useEffect(() => {
        if (!error) return;
        const timer = setTimeout(() => {
            router.replace('/');
        }, 5000);
        return () => clearTimeout(timer);
    }, [error]);

    // When no items are available, redirect to home after 5 seconds
    useEffect(() => {
        if (loading || error || playQueue.length > 0) return;
        const timer = setTimeout(() => {
            router.replace('/');
        }, 5000);
        return () => clearTimeout(timer);
    }, [loading, error, playQueue]);

    // Duration timer — depends ONLY on currentIndex (reads queue from ref)
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
    }, [currentIndex]);

    const advanceToNext = () => {
        setCurrentIndex(prev => {
            const queue = playQueueRef.current;
            if (prev + 1 >= queue.length) {
                router.replace('/');
                return prev;
            }
            return prev + 1;
        });
    };

    const goToPreviousItem = () => {
        setCurrentIndex((prev) => Math.max(prev - 1, 0));
    };

    const goToNextItem = () => {
        advanceToNext();
    };

    const currentItem = playQueue[currentIndex];

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

    return (
        <SafeAreaView style={styles.safe}>
            <View style={styles.topBar}>
                <Text style={styles.topTitle}>{headerTitle}</Text>
            </View>

            <View style={styles.container}>
                {(currentItem.module_type === "photo" || currentItem.module_type === "message") && (
                    <PhotoViewerScreen configuration={currentItem} />
                )}

                {(currentItem.module_type === "weather" ||
                currentItem.module_type === "sports" ||
                currentItem.module_type === "news" ||
                currentItem.module_type === "event") && (
                    <TextViewerScreen configuration={currentItem} />
                )}

                {currentItem.module_type === "video" && (
                    <GreetingScreen
                        configuration={currentItem}
                        onPlaybackComplete={goToNextItem}
                    />
                )}

                {currentItem.module_type === "audio" && (
                    <AudioScreen
                        configuration={currentItem}
                        onPlaybackComplete={goToNextItem}
                    />
                )}

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
// Copyright (C) 2026 mateika8814 (https://github.com/mateika8814)
// This file is part of allah-verge (fork of Clash Verge Rev).

import {
  AnalyticsRounded,
  CheckCircleOutlineRounded,
  ErrorOutlineRounded,
  PlayArrowRounded,
  RefreshRounded,
  RestartAltRounded,
} from '@mui/icons-material'
import {
  Box,
  Button,
  Chip,
  CircularProgress,
  IconButton,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Typography,
} from '@mui/material'
import dayjs from 'dayjs'
import React, { useCallback, useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'

import {
  DiagnosticsReport,
  FragmentStats,
  getFragmentStats,
  resetFragmentStats,
  runFragmentDiagnostics,
} from '@/services/cmds'
import { showNotice } from '@/services/notice-service'

export const FragmentDiagnosticsCard: React.FC = () => {
  const { t } = useTranslation()
  const [stats, setStats] = useState<FragmentStats | null>(null)
  const [report, setReport] = useState<DiagnosticsReport | null>(null)
  const [statsLoading, setStatsLoading] = useState(false)
  const [diagLoading, setDiagLoading] = useState(false)

  const loadStats = useCallback(async () => {
    setStatsLoading(true)
    try {
      const data = await getFragmentStats()
      setStats(data)
    } catch {
      // Core might be initializing or offline
    } finally {
      setStatsLoading(false)
    }
  }, [])

  useEffect(() => {
    void loadStats()
    const timer = setInterval(() => {
      void loadStats()
    }, 4000)
    return () => clearInterval(timer)
  }, [loadStats])

  const handleResetStats = async () => {
    try {
      await resetFragmentStats()
      await loadStats()
      showNotice.success(
        t('settings.sections.clash.form.fields.fragment.statsResetSuccess'),
        1500,
      )
    } catch (err: any) {
      showNotice.error(err)
    }
  }

  const handleRunDiagnostics = async () => {
    setDiagLoading(true)
    try {
      const res = await runFragmentDiagnostics()
      setReport(res)
    } catch (err: any) {
      showNotice.error(err)
    } finally {
      setDiagLoading(false)
    }
  }

  return (
    <Box
      sx={{
        mt: 2,
        pt: 2,
        borderTop: '1px solid',
        borderColor: (theme) =>
          theme.palette.mode === 'dark'
            ? 'rgba(255, 255, 255, 0.08)'
            : 'rgba(0, 0, 0, 0.08)',
        display: 'flex',
        flexDirection: 'column',
        gap: 2,
      }}
    >
      {/* Section 1: Live Statistics */}
      <Box>
        <Box
          sx={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            mb: 1.5,
          }}
        >
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
            <AnalyticsRounded fontSize="small" sx={{ color: '#00b0ff' }} />
            <Typography variant="subtitle2" sx={{ fontWeight: 600 }}>
              {t('settings.sections.clash.form.fields.fragment.statsTitle')}
            </Typography>
            {statsLoading && <CircularProgress size={14} sx={{ ml: 0.5 }} />}
          </Box>

          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
            <IconButton
              size="small"
              onClick={() => void loadStats()}
              title={t('settings.sections.clash.form.fields.fragment.statsRefresh')}
            >
              <RefreshRounded fontSize="small" />
            </IconButton>
            <Button
              size="small"
              variant="outlined"
              color="inherit"
              startIcon={<RestartAltRounded fontSize="small" />}
              onClick={() => void handleResetStats()}
              sx={{ borderRadius: '8px', textTransform: 'none', py: '2px', px: 1 }}
            >
              {t('settings.sections.clash.form.fields.fragment.statsReset')}
            </Button>
          </Box>
        </Box>

        {/* Counter Metric Cards */}
        <Box
          sx={{
            display: 'grid',
            gridTemplateColumns: { xs: 'repeat(2, 1fr)', sm: 'repeat(3, 1fr)' },
            gap: 1,
            mb: 1.5,
          }}
        >
          {/* Card: Total */}
          <Box
            sx={{
              p: 1.2,
              borderRadius: '10px',
              bgcolor: (theme) =>
                theme.palette.mode === 'dark'
                  ? 'rgba(255, 255, 255, 0.03)'
                  : 'rgba(0, 0, 0, 0.02)',
              border: '1px solid',
              borderColor: (theme) =>
                theme.palette.mode === 'dark'
                  ? 'rgba(255, 255, 255, 0.06)'
                  : 'rgba(0, 0, 0, 0.06)',
            }}
          >
            <Typography variant="caption" sx={{ color: 'text.secondary', display: 'block' }}>
              {t('settings.sections.clash.form.fields.fragment.totalClientHello')}
            </Typography>
            <Typography variant="h6" sx={{ fontWeight: 700, mt: 0.2 }}>
              {stats?.total_client_hello ?? 0}
            </Typography>
          </Box>

          {/* Card: Successful */}
          <Box
            sx={{
              p: 1.2,
              borderRadius: '10px',
              bgcolor: (theme) =>
                theme.palette.mode === 'dark'
                  ? 'rgba(46, 125, 50, 0.15)'
                  : 'rgba(76, 175, 80, 0.08)',
              border: '1px solid',
              borderColor: 'rgba(76, 175, 80, 0.2)',
            }}
          >
            <Typography variant="caption" sx={{ color: '#66bb6a', display: 'block' }}>
              {t('settings.sections.clash.form.fields.fragment.successfulFragments')}
            </Typography>
            <Typography variant="h6" sx={{ fontWeight: 700, color: '#66bb6a', mt: 0.2 }}>
              {stats?.successful_fragments ?? 0}
            </Typography>
          </Box>

          {/* Card: Errors */}
          <Box
            sx={{
              p: 1.2,
              borderRadius: '10px',
              bgcolor: (theme) =>
                theme.palette.mode === 'dark'
                  ? 'rgba(211, 47, 47, 0.15)'
                  : 'rgba(244, 67, 54, 0.08)',
              border: '1px solid',
              borderColor: 'rgba(244, 67, 54, 0.2)',
            }}
          >
            <Typography variant="caption" sx={{ color: '#ef5350', display: 'block' }}>
              {t('settings.sections.clash.form.fields.fragment.failedFragments')}
            </Typography>
            <Typography variant="h6" sx={{ fontWeight: 700, color: '#ef5350', mt: 0.2 }}>
              {stats?.failed_fragments ?? 0}
            </Typography>
          </Box>

          {/* Card: Timeouts */}
          <Box
            sx={{
              p: 1.2,
              borderRadius: '10px',
              bgcolor: (theme) =>
                theme.palette.mode === 'dark'
                  ? 'rgba(237, 108, 2, 0.15)'
                  : 'rgba(255, 152, 0, 0.08)',
              border: '1px solid',
              borderColor: 'rgba(255, 152, 0, 0.2)',
            }}
          >
            <Typography variant="caption" sx={{ color: '#ffa726', display: 'block' }}>
              {t('settings.sections.clash.form.fields.fragment.timeouts')}
            </Typography>
            <Typography variant="h6" sx={{ fontWeight: 700, color: '#ffa726', mt: 0.2 }}>
              {stats?.timeouts ?? 0}
            </Typography>
          </Box>

          {/* Card: Non-TLS Bypassed */}
          <Box
            sx={{
              p: 1.2,
              borderRadius: '10px',
              bgcolor: (theme) =>
                theme.palette.mode === 'dark'
                  ? 'rgba(2, 136, 209, 0.15)'
                  : 'rgba(3, 169, 244, 0.08)',
              border: '1px solid',
              borderColor: 'rgba(3, 169, 244, 0.2)',
            }}
          >
            <Typography variant="caption" sx={{ color: '#29b6f6', display: 'block' }}>
              {t('settings.sections.clash.form.fields.fragment.bypassedNonTls')}
            </Typography>
            <Typography variant="h6" sx={{ fontWeight: 700, color: '#29b6f6', mt: 0.2 }}>
              {stats?.bypassed_non_tls ?? 0}
            </Typography>
          </Box>

          {/* Card: Last Event */}
          <Box
            sx={{
              p: 1.2,
              borderRadius: '10px',
              bgcolor: (theme) =>
                theme.palette.mode === 'dark'
                  ? 'rgba(255, 255, 255, 0.03)'
                  : 'rgba(0, 0, 0, 0.02)',
              border: '1px solid',
              borderColor: (theme) =>
                theme.palette.mode === 'dark'
                  ? 'rgba(255, 255, 255, 0.06)'
                  : 'rgba(0, 0, 0, 0.06)',
            }}
          >
            <Typography variant="caption" sx={{ color: 'text.secondary', display: 'block' }}>
              {t('settings.sections.clash.form.fields.fragment.lastEvent')}
            </Typography>
            <Typography variant="body2" sx={{ fontWeight: 600, mt: 0.5 }}>
              {stats?.last_event_time
                ? dayjs.unix(stats.last_event_time).format('HH:mm:ss')
                : '-'}
            </Typography>
          </Box>
        </Box>

        {/* Mode Distribution Chips */}
        {stats?.mode_distribution && Object.keys(stats.mode_distribution).length > 0 && (
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, flexWrap: 'wrap', mb: 1 }}>
            <Typography variant="caption" sx={{ color: 'text.secondary' }}>
              {t('settings.sections.clash.form.fields.fragment.modeDistribution')}:
            </Typography>
            {Object.entries(stats.mode_distribution).map(([modeKey, count]) => (
              <Chip
                key={modeKey}
                size="small"
                label={`${modeKey.toUpperCase()}: ${count}`}
                sx={{
                  height: 22,
                  fontSize: '0.72rem',
                  fontWeight: 600,
                  bgcolor: 'rgba(124, 77, 255, 0.12)',
                  color: '#9c27b0',
                }}
              />
            ))}
          </Box>
        )}

        {/* Last Error banner if any */}
        {stats?.last_error && (
          <Box
            sx={{
              p: 1,
              borderRadius: '8px',
              bgcolor: 'rgba(244, 67, 54, 0.1)',
              border: '1px solid rgba(244, 67, 54, 0.3)',
              display: 'flex',
              alignItems: 'center',
              gap: 1,
            }}
          >
            <ErrorOutlineRounded fontSize="small" sx={{ color: '#ef5350' }} />
            <Typography variant="caption" sx={{ color: '#ef5350', fontWeight: 600 }}>
              {t('settings.sections.clash.form.fields.fragment.lastError')}: {stats.last_error}
              {stats.last_error_time
                ? ` (${dayjs.unix(stats.last_error_time).format('YYYY-MM-DD HH:mm:ss')})`
                : ''}
            </Typography>
          </Box>
        )}
      </Box>

      {/* Section 2: Diagnostics Runner */}
      <Box>
        <Box
          sx={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            mb: 1.5,
          }}
        >
          <Typography variant="subtitle2" sx={{ fontWeight: 600 }}>
            {t('settings.sections.clash.form.fields.fragment.diagTitle')}
          </Typography>

          <Button
            size="small"
            variant="contained"
            color="primary"
            startIcon={
              diagLoading ? <CircularProgress size={14} color="inherit" /> : <PlayArrowRounded />
            }
            disabled={diagLoading}
            onClick={() => void handleRunDiagnostics()}
            sx={{ borderRadius: '8px', textTransform: 'none', py: '3px', px: 1.5 }}
          >
            {diagLoading
              ? t('settings.sections.clash.form.fields.fragment.diagRunning')
              : t('settings.sections.clash.form.fields.fragment.diagRun')}
          </Button>
        </Box>

        {/* Diagnostics Report Summary */}
        {report && (
          <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1.5 }}>
            <Box
              sx={{
                p: 1.2,
                borderRadius: '8px',
                bgcolor: report.all_passed
                  ? 'rgba(76, 175, 80, 0.12)'
                  : 'rgba(244, 67, 54, 0.12)',
                border: '1px solid',
                borderColor: report.all_passed
                  ? 'rgba(76, 175, 80, 0.3)'
                  : 'rgba(244, 67, 54, 0.3)',
                display: 'flex',
                alignItems: 'center',
                gap: 1,
              }}
            >
              {report.all_passed ? (
                <CheckCircleOutlineRounded fontSize="small" sx={{ color: '#4caf50' }} />
              ) : (
                <ErrorOutlineRounded fontSize="small" sx={{ color: '#f44336' }} />
              )}
              <Typography
                variant="body2"
                sx={{
                  fontWeight: 600,
                  color: report.all_passed ? '#4caf50' : '#f44336',
                }}
              >
                {report.all_passed
                  ? t('settings.sections.clash.form.fields.fragment.diagAllPassed')
                  : t('settings.sections.clash.form.fields.fragment.diagSomeFailed')}{' '}
                (
                {report.checks.reduce((acc, c) => acc + c.duration_ms, 0)} ms,{' '}
                {dayjs.unix(report.timestamp).format('HH:mm:ss')})
              </Typography>
            </Box>

            {/* Checks Table */}
            <TableContainer
              sx={{
                borderRadius: '8px',
                border: '1px solid',
                borderColor: (theme) =>
                  theme.palette.mode === 'dark'
                    ? 'rgba(255, 255, 255, 0.08)'
                    : 'rgba(0, 0, 0, 0.08)',
              }}
            >
              <Table size="small">
                <TableHead>
                  <TableRow>
                    <TableCell sx={{ fontWeight: 600 }}>
                      {t('settings.sections.clash.form.fields.fragment.diagCheckName')}
                    </TableCell>
                    <TableCell sx={{ fontWeight: 600, width: 90 }}>
                      {t('settings.sections.clash.form.fields.fragment.diagStatus')}
                    </TableCell>
                    <TableCell sx={{ fontWeight: 600, width: 80 }}>
                      {t('settings.sections.clash.form.fields.fragment.diagDuration')}
                    </TableCell>
                    <TableCell sx={{ fontWeight: 600 }}>
                      {t('settings.sections.clash.form.fields.fragment.diagDetails')}
                    </TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {report.checks.map((chk) => (
                    <TableRow key={chk.name}>
                      <TableCell sx={{ fontWeight: 500 }}>{chk.name}</TableCell>
                      <TableCell>
                        <Chip
                          size="small"
                          label={chk.status}
                          sx={{
                            height: 20,
                            fontSize: '0.68rem',
                            fontWeight: 700,
                            bgcolor:
                              chk.status === 'PASS'
                                ? 'rgba(76, 175, 80, 0.15)'
                                : chk.status === 'FAIL'
                                  ? 'rgba(244, 67, 54, 0.15)'
                                  : 'rgba(158, 158, 158, 0.15)',
                            color:
                              chk.status === 'PASS'
                                ? '#4caf50'
                                : chk.status === 'FAIL'
                                  ? '#f44336'
                                  : '#9e9e9e',
                          }}
                        />
                      </TableCell>
                      <TableCell sx={{ color: 'text.secondary' }}>
                        {chk.duration_ms} ms
                      </TableCell>
                      <TableCell sx={{ color: 'text.secondary', fontSize: '0.8rem' }}>
                        {chk.description}
                        {chk.details ? ` (${chk.details})` : ''}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </TableContainer>
          </Box>
        )}
      </Box>
    </Box>
  )
}

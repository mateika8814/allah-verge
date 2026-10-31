import { DeleteForeverRounded, TuneRounded, UndoRounded } from '@mui/icons-material'
import {
  Box,
  IconButton,
  ListItem,
  ListItemText,
  Tooltip,
  alpha,
  styled,
} from '@mui/material'
import { useTranslation } from 'react-i18next'

import { isTlsProxy } from '@/utils/fragment'

interface Props {
  type: 'prepend' | 'original' | 'delete' | 'append'
  proxy: IProxyConfig
  onDelete: () => void
  onEditFragment?: () => void
}

export const ProxyItem = (props: Props) => {
  const { type, proxy, onDelete, onEditFragment } = props
  const { t } = useTranslation()
  const isSortable = type === 'prepend' || type === 'append'
  const isTls = isTlsProxy(proxy)
  const isFragmentEnabled = proxy.fragment?.enabled === true

  return (
    <ListItem
      dense
      sx={({ palette }) => ({
        position: 'relative',
        background:
          type === 'original'
            ? palette.mode === 'dark'
              ? alpha(palette.background.paper, 0.3)
              : alpha(palette.grey[400], 0.3)
            : type === 'delete'
              ? alpha(palette.error.main, 0.3)
              : alpha(palette.success.main, 0.3),
        height: '100%',
        borderRadius: '8px',
      })}
    >
      <ListItemText
        data-sortable-handle
        sx={{ cursor: isSortable ? 'move' : undefined }}
        primary={
          <StyledPrimary
            title={proxy.name}
            sx={{ textDecoration: type === 'delete' ? 'line-through' : '' }}
          >
            {proxy.name}
          </StyledPrimary>
        }
        secondary={
          <ListItemTextChild
            sx={{
              overflow: 'hidden',
              display: 'flex',
              alignItems: 'center',
              pt: '2px',
              gap: 0.5,
            }}
          >
            <Box sx={{ marginTop: '2px' }}>
              <StyledTypeBox>{proxy.type}</StyledTypeBox>
            </Box>
            {isFragmentEnabled && (
              <Box sx={{ marginTop: '2px' }}>
                <StyledFragBadge>
                  {proxy.fragment?.mode === 'random' ? 'FRAG:RAND' : 'FRAG:SNI'}
                </StyledFragBadge>
              </Box>
            )}
          </ListItemTextChild>
        }
        slotProps={{
          secondary: {
            sx: {
              display: 'flex',
              alignItems: 'center',
              color: '#ccc',
            },
          },
        }}
      />

      {isTls && onEditFragment && type !== 'delete' && (
        <Tooltip title={t('profiles.modals.proxiesEditor.fragmentDialog.title')}>
          <IconButton
            size="small"
            onClick={onEditFragment}
            sx={{
              mr: 0.5,
              color: isFragmentEnabled ? '#66bb6a' : 'text.secondary',
            }}
          >
            <TuneRounded fontSize="small" />
          </IconButton>
        </Tooltip>
      )}

      <IconButton onClick={onDelete}>
        {type === 'delete' ? <UndoRounded /> : <DeleteForeverRounded />}
      </IconButton>
    </ListItem>
  )
}

const StyledPrimary = styled('div')`
  font-size: 15px;
  font-weight: 700;
  line-height: 1.5;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
`

const ListItemTextChild = styled('span')`
  display: block;
`

const StyledTypeBox = styled(ListItemTextChild)(({ theme }) => ({
  display: 'inline-block',
  border: '1px solid #ccc',
  borderColor: alpha(theme.palette.primary.main, 0.5),
  color: alpha(theme.palette.primary.main, 0.8),
  borderRadius: 4,
  fontSize: 10,
  padding: '0 4px',
}))

const StyledFragBadge = styled(ListItemTextChild)(() => ({
  display: 'inline-block',
  border: '1px solid #66bb6a',
  borderColor: alpha('#66bb6a', 0.6),
  color: '#66bb6a',
  borderRadius: 4,
  fontSize: 9,
  fontWeight: 700,
  padding: '0 4px',
}))
